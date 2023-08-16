// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	humanize "github.com/dustin/go-humanize"
	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio/internal/logger"
	"github.com/minio/mux"
	"github.com/minio/pkg/bucket/policy"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

const (
	// As per AWS S3 specification, 20KiB policy JSON data is allowed.
	maxBucketPolicySize = 20 * humanize.KiByte

	// Policy configuration file.
	bucketPolicyConfig = "policy.json"
)

// PutBucketPolicyHandler - This HTTP handler stores given bucket policy configuration as per
// https://docs.aws.amazon.com/AmazonS3/latest/dev/access-policy-language-overview.html
func (api objectAPIHandlers) PutBucketPolicyHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Received PutBucketPolicy request", zap.String("client", r.RemoteAddr), zap.Any("headers", r.Header))

	ctx := newContext(r, w, "PutBucketPolicy")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objAPI := api.ObjectAPI()
	if objAPI == nil {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(ErrServerNotInitialized), r.URL)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if s3Error := checkRequestAuthType(ctx, r, policy.PutBucketPolicyAction, bucket, ""); s3Error != ErrNone {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(s3Error), r.URL)
		return
	}

	// Check if bucket exists.
	if _, err := objAPI.GetBucketInfo(ctx, bucket, BucketOptions{}); err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	// Error out if Content-Length is missing.
	// PutBucketPolicy always needs Content-Length.
	if r.ContentLength <= 0 {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(ErrMissingContentLength), r.URL)
		return
	}

	// Error out if Content-Length is beyond allowed size.
	if r.ContentLength > maxBucketPolicySize {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(ErrPolicyTooLarge), r.URL)
		return
	}

	bucketPolicyBytes, err := io.ReadAll(io.LimitReader(r.Body, r.ContentLength))
	if err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	bucketPolicy, err := policy.ParseConfig(bytes.NewReader(bucketPolicyBytes), bucket)
	if err != nil {
		writeErrorResponse(ctx, w, APIError{
			Code:           "MalformedPolicy",
			HTTPStatusCode: http.StatusBadRequest,
			Description:    err.Error(),
		}, r.URL)
		return
	}

	extractObjectNames := func(bucketPolicy *policy.Policy) []string {
		var objectNames []string
		for _, statement := range bucketPolicy.Statements {
			if statement.Effect == "Allow" {
				for principal := range statement.Principal.AWS {
					if principal == "*" {
						if _, ok := statement.Actions[policy.GetObjectAction]; ok {
							for resource := range statement.Resources {
								objectName := resource.Pattern
								// Remove the bucket prefix if present
								if strings.Contains(objectName, "/") {
									objectName = strings.SplitN(objectName, "/", 2)[1]
								} else if objectName == bucket {
									// If the resource is the bucket itself, skip it
									continue
								}
								objectNames = append(objectNames, objectName)
							}
						}
					}
				}
			}
		}
		return objectNames
	}

	// Read bucket access policy.
	var originalBucketPolicy *policy.Policy
	modelKey := fmt.Sprintf("%s-%s-%s", api.DidManagerId, "minio_bucket_policy", bucket)
	modelExists := false
	var content []byte
	modelResponse, err := api.SaoClient.GetModel(ctx, modelKey)
	if err == nil {
		dataId := modelResponse.Model.Data
		content, err = api.SaoClient.Load(ctx, dataId, "", "", bucket)
		if err == nil {
			err = json.Unmarshal(content, &originalBucketPolicy)
			modelExists = true
		}
	}

	// If the model doesn't exist, read the policy from local server
	if err != nil {
		logger.Info("Unable to read original bucket policy from SAO", zap.Error(err))
		if strings.Contains(err.Error(), "no route to host") {
			writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
			return
		}

		originalBucketPolicy, err = globalPolicySys.Get(bucket)
		if err != nil {
			logger.Error("Unable to read original bucket policy", zap.Error(err))
			newObjectNames := extractObjectNames(bucketPolicy)
			api.updateObjectPermissions(ctx, newObjectNames, bucket, true)
		}
	}

	if originalBucketPolicy != nil {
		// Extract object names from a policy
		originalObjectNames := extractObjectNames(originalBucketPolicy)
		newObjectNames := extractObjectNames(bucketPolicy)

		// Find added and removed object names
		addedObjectNames := difference(newObjectNames, originalObjectNames)
		removedObjectNames := difference(originalObjectNames, newObjectNames)

		api.updateObjectPermissions(ctx, addedObjectNames, bucket, true)
		api.updateObjectPermissions(ctx, removedObjectNames, bucket, false)
	}

	// Version in policy must not be empty
	if bucketPolicy.Version == "" {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(ErrPolicyInvalidVersion), r.URL)
		return
	}

	configData, err := json.Marshal(bucketPolicy)
	if err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	// Marshal the bucket policy to JSON
	jsonData, err := json.Marshal(bucketPolicy)
	if err != nil {
		logger.Error("Error marshaling bucket policy", zap.Error(err))
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	if modelExists {
		// Model exists, update it

		// print jsonData
		logger.Info("jsonData", zap.String("jsonData", string(jsonData)))
		// print content
		logger.Info("original content", zap.String("content", string(content)))

		//print modelResponse.Model.Data
		logger.Info("modelResponse.Model.Data", zap.String("modelResponse.Model.Data", modelResponse.Model.Data))
		err := api.SaoClient.UpdateModelQuick(ctx, modelResponse.Model.Data, jsonData, bucket, 365, 30, false, 1)
		if err != nil {
			if strings.Contains(err.Error(), "No differences found") {
				logger.Info("No differences found, model not updated")
			} else {
				logger.Error("Error updating model for bucket policy", zap.Error(err))
				writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
				return
			}
		} else {
			logger.Info("Bucket policy model updated")
		}
	} else {
		// Create a new model for the bucket policy using the SAO client
		_, dataId, err := api.SaoClient.CreateModel(ctx, string(jsonData), bucket, 365, 30, "minio_bucket_policy", 1, false)
		if err != nil {
			logger.Error("Error creating model for bucket policy", zap.Error(err))
			writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
			return
		}
		logger.Info("Bucket policy model created", zap.String("dataId", dataId))
	}

	updatedAt, err := globalBucketMetadataSys.Update(ctx, bucket, bucketPolicyConfig, configData)
	if err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	// Call site replication hook.
	logger.LogIf(ctx, globalSiteReplicationSys.BucketMetaHook(ctx, madmin.SRBucketMeta{
		Type:      madmin.SRBucketMetaTypePolicy,
		Bucket:    bucket,
		Policy:    bucketPolicyBytes,
		UpdatedAt: updatedAt,
	}))

	// Success.
	writeSuccessNoContent(w)
}

func (api objectAPIHandlers) updateObjectPermissions(ctx context.Context, objectNames []string, bucket string, addPermission bool) {
	if !addPermission && contains(objectNames, "*") {
		logger.Info("Don't remove public read access from all objects in bucket")
		return
	}

nextObjectName:
	for _, objectName := range objectNames {
		if objectName == "*" {
			logger.Info("* is not supported")
			continue
		}

		if addPermission {
			logger.Info("New object made publicly readable: %s\n", objectName)
		} else {
			logger.Info("Object removed from public read access: %s\n", objectName)
		}

		for _, suffix := range []string{"file_" + objectName, objectName + "_info"} {
			dataId, err := api.fetchSaoDataId(ctx, api.DidManagerId, suffix, bucket)
			if err != nil {
				logger.Error("fetchSaoDataId error: %s\n", err.Error())
				continue nextObjectName
			}

			if addPermission {
				err = api.SaoClient.SetPublicPermission(ctx, dataId)
				if err != nil {
					logger.Error("SetPublicPermission error: %s\n", err.Error())
					continue nextObjectName
				}
			} else {
				err = api.SaoClient.UpdatePermission(ctx, dataId, []string{}, []string{})
				if err != nil {
					logger.Error("SetPublicPermission error: %s\n", err.Error())
					continue nextObjectName
				}
			}
		}
	}
}

func (api objectAPIHandlers) fetchSaoDataId(ctx context.Context, didManagerId, object, bucket string) (string, error) {
	modelKey := fmt.Sprintf("%s-%s-%s", didManagerId, object, bucket)
	logger.Info("modelKey: ", modelKey)
	// Call saoClient.GetModel() to get the dataId
	modelResponse, err := api.SaoClient.GetModel(ctx, modelKey)
	if err != nil {
		logger.Error("Failed to fetch sao data Id", zap.Error(err))
		return "", err
	}

	logger.Info(modelResponse.Model.Data)

	// Return the dataId from the modelResponse
	return modelResponse.Model.Data, nil
}

// Utility function to check if a slice contains a specific string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func difference(slice1, slice2 []string) []string {
	var diff []string
	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}
	return diff
}

// DeleteBucketPolicyHandler - This HTTP handler removes bucket policy configuration.
func (api objectAPIHandlers) DeleteBucketPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "DeleteBucketPolicy")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objAPI := api.ObjectAPI()
	if objAPI == nil {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(ErrServerNotInitialized), r.URL)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if s3Error := checkRequestAuthType(ctx, r, policy.DeleteBucketPolicyAction, bucket, ""); s3Error != ErrNone {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(s3Error), r.URL)
		return
	}

	// Check if bucket exists.
	if _, err := objAPI.GetBucketInfo(ctx, bucket, BucketOptions{}); err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	updatedAt, err := globalBucketMetadataSys.Delete(ctx, bucket, bucketPolicyConfig)
	if err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	// Call site replication hook.
	logger.LogIf(ctx, globalSiteReplicationSys.BucketMetaHook(ctx, madmin.SRBucketMeta{
		Type:      madmin.SRBucketMetaTypePolicy,
		Bucket:    bucket,
		UpdatedAt: updatedAt,
	}))

	// Success.
	writeSuccessNoContent(w)
}

// GetBucketPolicyHandler - This HTTP handler returns bucket policy configuration.
func (api objectAPIHandlers) GetBucketPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "GetBucketPolicy")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objAPI := api.ObjectAPI()
	if objAPI == nil {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(ErrServerNotInitialized), r.URL)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if s3Error := checkRequestAuthType(ctx, r, policy.GetBucketPolicyAction, bucket, ""); s3Error != ErrNone {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(s3Error), r.URL)
		return
	}

	// Check if bucket exists.
	if _, err := objAPI.GetBucketInfo(ctx, bucket, BucketOptions{}); err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	// Read bucket access policy.
	config, err := globalPolicySys.Get(bucket)
	if err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	configData, err := json.Marshal(config)
	if err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}

	// Write to client.
	writeSuccessResponseJSON(w, configData)
}
