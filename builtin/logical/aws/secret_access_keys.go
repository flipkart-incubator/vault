// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package aws

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-secure-stdlib/awsutil"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/template"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	secretAccessKeyType = "access_keys"
	storageKey          = "config/root"
)

func secretAccessKeys(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: secretAccessKeyType,
		Fields: map[string]*framework.FieldSchema{
			"access_key": {
				Type:        framework.TypeString,
				Description: "Access Key",
			},

			"secret_key": {
				Type:        framework.TypeString,
				Description: "Secret Key",
			},
			"session_token": {
				Type:        framework.TypeString,
				Description: "Session Token",
			},
			"security_token": {
				Type:        framework.TypeString,
				Description: "Security Token",
				Deprecated:  true,
			},
		},

		Renew:  b.secretAccessKeysRenew,
		Revoke: b.secretAccessKeysRevoke,
	}
}

func genUsername(displayName, policyName, userType, usernameTemplate string) (ret string, err error) {
	switch userType {
	case "iam_user", "assume_role":
		// IAM users are capped at 64 chars
		up, err := template.NewTemplate(template.Template(usernameTemplate))
		if err != nil {
			return "", fmt.Errorf("unable to initialize username template: %w", err)
		}

		um := UsernameMetadata{
			Type:        "IAM",
			DisplayName: normalizeDisplayName(displayName),
			PolicyName:  normalizeDisplayName(policyName),
		}

		ret, err = up.Generate(um)
		if err != nil {
			return "", fmt.Errorf("failed to generate username: %w", err)
		}
		// To prevent a custom template from exceeding IAM length limits
		if len(ret) > 64 {
			return "", fmt.Errorf("the username generated by the template exceeds the IAM username length limits of 64 chars")
		}
	case "sts":
		up, err := template.NewTemplate(template.Template(usernameTemplate))
		if err != nil {
			return "", fmt.Errorf("unable to initialize username template: %w", err)
		}

		um := UsernameMetadata{
			Type: "STS",
		}
		ret, err = up.Generate(um)
		if err != nil {
			return "", fmt.Errorf("failed to generate username: %w", err)
		}
		// To prevent a custom template from exceeding STS length limits
		if len(ret) > 32 {
			return "", fmt.Errorf("the username generated by the template exceeds the STS username length limits of 32 chars")
		}
	}
	return
}

func (b *backend) getFederationToken(ctx context.Context, s logical.Storage,
	displayName, policyName, policy string, policyARNs []string,
	iamGroups []string, lifeTimeInSeconds int64) (*logical.Response, error,
) {
	groupPolicies, groupPolicyARNs, err := b.getGroupPolicies(ctx, s, iamGroups)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}
	if groupPolicies != nil {
		groupPolicies = append(groupPolicies, policy)
		policy, err = combinePolicyDocuments(groupPolicies...)
		if err != nil {
			return logical.ErrorResponse(err.Error()), nil
		}
	}
	if len(groupPolicyARNs) > 0 {
		policyARNs = append(policyARNs, groupPolicyARNs...)
	}

	stsClient, err := b.clientSTS(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	config, err := readConfig(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("unable to read configuration: %w", err)
	}

	// Set as defaultUsernameTemplate if not provided
	usernameTemplate := config.UsernameTemplate
	if usernameTemplate == "" {
		usernameTemplate = defaultUserNameTemplate
	}

	username, usernameError := genUsername(displayName, policyName, "sts", usernameTemplate)
	// Send a 400 to Framework.OperationFunc Handler
	if usernameError != nil {
		return nil, usernameError
	}

	getTokenInput := &sts.GetFederationTokenInput{
		Name:            aws.String(username),
		DurationSeconds: &lifeTimeInSeconds,
	}
	if len(policy) > 0 {
		getTokenInput.Policy = aws.String(policy)
	}
	if len(policyARNs) > 0 {
		getTokenInput.PolicyArns = convertPolicyARNs(policyARNs)
	}

	// If neither a policy document nor policy ARNs are specified, then GetFederationToken will
	// return credentials equivalent to that of the Vault server itself. We probably don't want
	// that by default; the behavior can be explicitly opted in to by associating the Vault role
	// with a policy ARN or document that allows the appropriate permissions.
	if policy == "" && len(policyARNs) == 0 {
		return logical.ErrorResponse("must specify at least one of policy_arns or policy_document with %s credential_type", federationTokenCred), nil
	}

	tokenResp, err := stsClient.GetFederationTokenWithContext(ctx, getTokenInput)
	if err != nil {
		return logical.ErrorResponse("Error generating STS keys: %s", err), awsutil.CheckAWSError(err)
	}

	// While STS credentials cannot be revoked/renewed, we will still create a lease since users are
	// relying on a non-zero `lease_duration` in order to manage their lease lifecycles manually.
	//
	ttl := time.Until(*tokenResp.Credentials.Expiration)
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		"access_key":     *tokenResp.Credentials.AccessKeyId,
		"secret_key":     *tokenResp.Credentials.SecretAccessKey,
		"security_token": *tokenResp.Credentials.SessionToken,
		"session_token":  *tokenResp.Credentials.SessionToken,
		"ttl":            uint64(ttl.Seconds()),
	}, map[string]interface{}{
		"username": username,
		"policy":   policy,
		"is_sts":   true,
	})

	// Set the secret TTL to appropriately match the expiration of the token
	resp.Secret.TTL = ttl

	// STS are purposefully short-lived and aren't renewable
	resp.Secret.Renewable = false

	return resp, nil
}

// NOTE: Getting session tokens with or without MFA/TOTP has behavior that can cause confusion.
// When an AWS IAM user has a policy attached requiring an MFA code by use of "aws:MultiFactorAuthPresent": "true",
// then credentials may still be returned without an MFA code provided.
// If a Vault role associated with the IAM user is configured without both an mfa_serial_number and
// the mfa_code is not given, the API call is successful and returns credentials. These credentials
// are scoped to any resources in the policy that do NOT have "aws:MultiFactorAuthPresent": "true" set and
// accessing resources with it set will be denied.
// This is expected behavior, as the policy may have a mix of permissions, some requiring MFA and others not.
// If an mfa_serial_number is set on the Vault role, then a valid mfa_code MUST be provided to succeed.
func (b *backend) getSessionToken(ctx context.Context, s logical.Storage, serialNumber, mfaCode string, lifeTimeInSeconds int64) (*logical.Response, error) {
	stsClient, err := b.clientSTS(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	getTokenInput := &sts.GetSessionTokenInput{
		DurationSeconds: &lifeTimeInSeconds,
	}
	if serialNumber != "" {
		getTokenInput.SerialNumber = &serialNumber
	}
	if mfaCode != "" {
		getTokenInput.TokenCode = &mfaCode
	}

	tokenResp, err := stsClient.GetSessionToken(getTokenInput)
	if err != nil {
		return logical.ErrorResponse("Error generating STS keys: %s", err), awsutil.CheckAWSError(err)
	}

	ttl := time.Until(*tokenResp.Credentials.Expiration)
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		"access_key":    *tokenResp.Credentials.AccessKeyId,
		"secret_key":    *tokenResp.Credentials.SecretAccessKey,
		"session_token": *tokenResp.Credentials.SessionToken,
		"ttl":           uint64(ttl.Seconds()),
	}, map[string]interface{}{
		"is_sts": true,
	})

	// Set the secret TTL to appropriately match the expiration of the token
	resp.Secret.TTL = time.Until(*tokenResp.Credentials.Expiration)

	// STS are purposefully short-lived and aren't renewable
	resp.Secret.Renewable = false

	return resp, nil
}

func (b *backend) assumeRole(ctx context.Context, s logical.Storage,
	displayName, roleName, roleArn, policy string, policyARNs []string,
	iamGroups []string, lifeTimeInSeconds int64, roleSessionName string) (*logical.Response, error,
) {
	// grab any IAM group policies associated with the vault role, both inline
	// and managed
	groupPolicies, groupPolicyARNs, err := b.getGroupPolicies(ctx, s, iamGroups)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}
	if len(groupPolicies) > 0 {
		groupPolicies = append(groupPolicies, policy)
		policy, err = combinePolicyDocuments(groupPolicies...)
		if err != nil {
			return logical.ErrorResponse(err.Error()), nil
		}
	}
	if len(groupPolicyARNs) > 0 {
		policyARNs = append(policyARNs, groupPolicyARNs...)
	}

	stsClient, err := b.clientSTS(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	config, err := readConfig(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("unable to read configuration: %w", err)
	}

	// Set as defaultUsernameTemplate if not provided
	usernameTemplate := config.UsernameTemplate
	if usernameTemplate == "" {
		usernameTemplate = defaultUserNameTemplate
	}

	var roleSessionNameError error
	if roleSessionName == "" {
		roleSessionName, roleSessionNameError = genUsername(displayName, roleName, "assume_role", usernameTemplate)
		// Send a 400 to Framework.OperationFunc Handler
		if roleSessionNameError != nil {
			return nil, roleSessionNameError
		}
	} else {
		roleSessionName = normalizeDisplayName(roleSessionName)
	}

	assumeRoleInput := &sts.AssumeRoleInput{
		RoleSessionName: aws.String(roleSessionName),
		RoleArn:         aws.String(roleArn),
		DurationSeconds: &lifeTimeInSeconds,
	}
	if policy != "" {
		assumeRoleInput.SetPolicy(policy)
	}
	if len(policyARNs) > 0 {
		assumeRoleInput.SetPolicyArns(convertPolicyARNs(policyARNs))
	}
	tokenResp, err := stsClient.AssumeRoleWithContext(ctx, assumeRoleInput)
	if err != nil {
		return logical.ErrorResponse("Error assuming role: %s", err), awsutil.CheckAWSError(err)
	}

	// While STS credentials cannot be revoked/renewed, we will still create a lease since users are
	// relying on a non-zero `lease_duration` in order to manage their lease lifecycles manually.
	//
	ttl := time.Until(*tokenResp.Credentials.Expiration)
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		"access_key":     *tokenResp.Credentials.AccessKeyId,
		"secret_key":     *tokenResp.Credentials.SecretAccessKey,
		"security_token": *tokenResp.Credentials.SessionToken,
		"session_token":  *tokenResp.Credentials.SessionToken,
		"arn":            *tokenResp.AssumedRoleUser.Arn,
		"ttl":            uint64(ttl.Seconds()),
	}, map[string]interface{}{
		"username": roleSessionName,
		"policy":   roleArn,
		"is_sts":   true,
	})

	// Set the secret TTL to appropriately match the expiration of the token
	resp.Secret.TTL = ttl

	// STS are purposefully short-lived and aren't renewable
	resp.Secret.Renewable = false

	return resp, nil
}

func readConfig(ctx context.Context, storage logical.Storage) (rootConfig, error) {
	entry, err := storage.Get(ctx, storageKey)
	if err != nil {
		return rootConfig{}, err
	}
	if entry == nil {
		return rootConfig{}, nil
	}

	var connConfig rootConfig
	if err := entry.DecodeJSON(&connConfig); err != nil {
		return rootConfig{}, err
	}
	return connConfig, nil
}

func (b *backend) secretAccessKeysCreate(
	ctx context.Context,
	s logical.Storage,
	displayName, policyName string,
	role *awsRoleEntry,
) (*logical.Response, error) {
	iamClient, err := b.clientIAM(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	config, err := readConfig(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("unable to read configuration: %w", err)
	}

	// Set as defaultUsernameTemplate if not provided
	usernameTemplate := config.UsernameTemplate
	if usernameTemplate == "" {
		usernameTemplate = defaultUserNameTemplate
	}

	username, usernameError := genUsername(displayName, policyName, "iam_user", usernameTemplate)
	// Send a 400 to Framework.OperationFunc Handler
	if usernameError != nil {
		return nil, usernameError
	}

	// Write to the WAL that this user will be created. We do this before
	// the user is created because if switch the order then the WAL put
	// can fail, which would put us in an awkward position: we have a user
	// we need to rollback but can't put the WAL entry to do the rollback.
	walID, err := framework.PutWAL(ctx, s, "user", &walUser{
		UserName: username,
	})
	if err != nil {
		return nil, fmt.Errorf("error writing WAL entry: %w", err)
	}

	userPath := role.UserPath
	if userPath == "" {
		userPath = "/"
	}

	createUserRequest := &iam.CreateUserInput{
		UserName: aws.String(username),
		Path:     aws.String(userPath),
	}
	if role.PermissionsBoundaryARN != "" {
		createUserRequest.PermissionsBoundary = aws.String(role.PermissionsBoundaryARN)
	}

	// Create the user
	_, err = iamClient.CreateUserWithContext(ctx, createUserRequest)
	if err != nil {
		if walErr := framework.DeleteWAL(ctx, s, walID); walErr != nil {
			iamErr := fmt.Errorf("error creating IAM user: %w", err)
			return nil, errwrap.Wrap(fmt.Errorf("failed to delete WAL entry: %w", walErr), iamErr)
		}
		return logical.ErrorResponse("Error creating IAM user: %s", err), awsutil.CheckAWSError(err)
	}

	for _, arn := range role.PolicyArns {
		// Attach existing policy against user
		_, err = iamClient.AttachUserPolicyWithContext(ctx, &iam.AttachUserPolicyInput{
			UserName:  aws.String(username),
			PolicyArn: aws.String(arn),
		})
		if err != nil {
			return logical.ErrorResponse("Error attaching user policy: %s", err), awsutil.CheckAWSError(err)
		}

	}
	if role.PolicyDocument != "" {
		// Add new inline user policy against user
		_, err = iamClient.PutUserPolicyWithContext(ctx, &iam.PutUserPolicyInput{
			UserName:       aws.String(username),
			PolicyName:     aws.String(policyName),
			PolicyDocument: aws.String(role.PolicyDocument),
		})
		if err != nil {
			return logical.ErrorResponse("Error putting user policy: %s", err), awsutil.CheckAWSError(err)
		}
	}

	for _, group := range role.IAMGroups {
		// Add user to IAM groups
		_, err = iamClient.AddUserToGroupWithContext(ctx, &iam.AddUserToGroupInput{
			UserName:  aws.String(username),
			GroupName: aws.String(group),
		})
		if err != nil {
			return logical.ErrorResponse("Error adding user to group: %s", err), awsutil.CheckAWSError(err)
		}
	}

	var tags []*iam.Tag
	for key, value := range role.IAMTags {
		// This assignment needs to be done in order to create unique addresses for
		// these variables. Without doing so, all the tags will be copies of the last
		// tag listed in the role.
		k, v := key, value
		tags = append(tags, &iam.Tag{Key: &k, Value: &v})
	}

	if len(tags) > 0 {
		_, err = iamClient.TagUserWithContext(ctx, &iam.TagUserInput{
			Tags:     tags,
			UserName: &username,
		})
		if err != nil {
			return logical.ErrorResponse("Error adding tags to user: %s", err), awsutil.CheckAWSError(err)
		}
	}

	// Create the keys
	keyResp, err := iamClient.CreateAccessKeyWithContext(ctx, &iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return logical.ErrorResponse("Error creating access keys: %s", err), awsutil.CheckAWSError(err)
	}

	// Remove the WAL entry, we succeeded! If we fail, we don't return
	// the secret because it'll get rolled back anyways, so we have to return
	// an error here.
	if err := framework.DeleteWAL(ctx, s, walID); err != nil {
		return nil, fmt.Errorf("failed to commit WAL entry: %w", err)
	}

	// Return the info!
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		"access_key":    *keyResp.AccessKey.AccessKeyId,
		"secret_key":    *keyResp.AccessKey.SecretAccessKey,
		"session_token": nil,
	}, map[string]interface{}{
		"username": username,
		"policy":   role,
		"is_sts":   false,
	})

	lease, err := b.Lease(ctx, s)
	if err != nil || lease == nil {
		lease = &configLease{}
	}

	resp.Secret.TTL = lease.Lease
	resp.Secret.MaxTTL = lease.LeaseMax

	return resp, nil
}

func (b *backend) secretAccessKeysRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// STS already has a lifetime, and we don't support renewing it
	isSTSRaw, ok := req.Secret.InternalData["is_sts"]
	if ok {
		isSTS, ok := isSTSRaw.(bool)
		if ok {
			if isSTS {
				return nil, nil
			}
		}
	}

	lease, err := b.Lease(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if lease == nil {
		lease = &configLease{}
	}

	resp := &logical.Response{Secret: req.Secret}
	resp.Secret.TTL = lease.Lease
	resp.Secret.MaxTTL = lease.LeaseMax
	return resp, nil
}

func (b *backend) secretAccessKeysRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// STS cleans up after itself so we can skip this if is_sts internal data
	// element set to true. If is_sts is not set, assumes old version
	// and defaults to the IAM approach.
	isSTSRaw, ok := req.Secret.InternalData["is_sts"]
	if ok {
		isSTS, ok := isSTSRaw.(bool)
		if ok {
			if isSTS {
				return nil, nil
			}
		} else {
			return nil, fmt.Errorf("secret has is_sts but value could not be understood")
		}
	}

	// Get the username from the internal data
	usernameRaw, ok := req.Secret.InternalData["username"]
	if !ok {
		return nil, fmt.Errorf("secret is missing username internal data")
	}
	username, ok := usernameRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing username internal data")
	}

	// Use the user rollback mechanism to delete this user
	err := b.pathUserRollback(ctx, req, "user", map[string]interface{}{
		"username": username,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func normalizeDisplayName(displayName string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9+=,.@_-]")
	return re.ReplaceAllString(displayName, "_")
}

func convertPolicyARNs(policyARNs []string) []*sts.PolicyDescriptorType {
	size := len(policyARNs)
	retval := make([]*sts.PolicyDescriptorType, size, size)
	for i, arn := range policyARNs {
		retval[i] = &sts.PolicyDescriptorType{
			Arn: aws.String(arn),
		}
	}
	return retval
}

type UsernameMetadata struct {
	Type        string
	DisplayName string
	PolicyName  string
}
