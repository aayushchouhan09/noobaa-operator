package validations

import (
	"encoding/json"
	"fmt"
	"regexp"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

var linuxUsernameRegex = regexp.MustCompile(`^[^-:\s][^\s:]{0,30}[^-:\s]$`)

// ValidateNSFSAccountConfig validates that the provided NSFS config is valid
func ValidateNSFSAccountConfig(NSFSConfig string, bucketclass string) error {
	log := util.Logger()

	if NSFSConfig == "" {
		return nil
	}

	var configObj nbv1.AccountNsfsConfig

	err := json.Unmarshal([]byte(NSFSConfig), &configObj)
	if err != nil {
		return fmt.Errorf("failed to parse NSFS account config %q: %v", NSFSConfig, err)
	}

	log.Infof("Validating NSFS config: %+v", NSFSConfig)
	if bucketclass == "" {
		return fmt.Errorf("a bucketclass backed by an NSFS namespacestore is required for NSFS account config usage")
	}

	// Require both UID and GID together, or DistinguishedName alone
	if (configObj.UID == nil || configObj.GID == nil) && configObj.DistinguishedName == "" {
		return fmt.Errorf("UID and GID, or DistinguishedName must be provided")
	}
	// Check UID/GID cases when numeric identity mapping is used
	if configObj.UID != nil || configObj.GID != nil {
		if configObj.UID == nil || configObj.GID == nil {
			return fmt.Errorf("NSFS account config must include both UID and GID")
		}
		if *configObj.UID < 0 || *configObj.GID < 0 {
			return fmt.Errorf("UID and GID must be positive integers")
		} else if configObj.DistinguishedName != "" {
			// Distinguished name cannot be provided alongside UID/GID
			return fmt.Errorf(`NSFS account config cannot include both distinguished name and UID/GID`)
		}
	} else if configObj.DistinguishedName != "" {
		// Validate distinguished name when used without UID/GID
		if !linuxUsernameRegex.MatchString(configObj.DistinguishedName) {
			return fmt.Errorf("DistinguishedName must be a valid username by Linux standards")
		}
	}

	return nil
}

// ValidateReplicationPolicy validates and replication params and returns the replication policy object
func ValidateReplicationPolicy(bucketName string, replicationPolicy string, update bool, isCLI bool) error {
	log := util.Logger()
	if replicationPolicy == "" {
		return nil
	}

	var replicationRules nb.ReplicationPolicy
	err := json.Unmarshal([]byte(replicationPolicy), &replicationRules)
	if err != nil {
		return fmt.Errorf("Failed to parse replication json %q: %v", replicationRules, err)
	}
	log.Infof("ValidateReplicationPolicy: newReplication %+v", replicationRules)

	if len(replicationRules.Rules) == 0 {
		if update {
			return nil
		}
		return fmt.Errorf("replication rules array of bucket %q is empty %q", bucketName, replicationRules)
	}

	replicationParams := &nb.BucketReplicationParams{
		Name:              bucketName,
		ReplicationPolicy: replicationRules,
	}

	log.Infof("ValidateReplicationPolicy: validating replication: replicationParams: %+v", replicationParams)

	var sysClient *system.Client
	if isCLI {
		sysClient, err = system.ConnectAuto()
	} else {
		sysClient, err = system.Connect(util.IsTestEnv())
	}
	if err != nil {
		return fmt.Errorf("Provisioner Failed to validate replication of bucket %q with error: %v", bucketName, err)
	}

	err = sysClient.NBClient.ValidateReplicationAPI(*replicationParams)
	if err != nil {
		rpcErr, isRPCErr := err.(*nb.RPCError)
		if isRPCErr {
			if rpcErr.RPCCode == "INVALID_REPLICATION_POLICY" {
				return fmt.Errorf("Bucket replication configuration is invalid")
			}
			if rpcErr.RPCCode == "INVALID_LOG_REPLICATION_INFO" {
				return fmt.Errorf("Bucket log replication info configuration is invalid")
			}
		}
		return fmt.Errorf("Provisioner Failed to validate replication of bucket %q with error: %v", bucketName, err)
	}
	log.Infof("ValidateReplicationPolicy: validated replication successfully")
	return nil
}
