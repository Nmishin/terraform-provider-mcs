package mcs

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Datastore names
const (
	Redis       = "redis"
	MongoDB     = "mongodb"
	PostgresPro = "postgrespro"
	Galera      = "galera_mysql"
	Postgres    = "postgresql"
	Clickhouse  = "clickhouse"
	MySQL       = "mysql"
)

func getClusterDatastores() []string {
	return []string{Galera, Postgres}
}

func getClusterWithShardsDatastores() []string {
	return []string{Clickhouse}
}

func extractDatabaseDatastore(v []interface{}) (dataStore, error) {
	var D dataStore
	in := v[0].(map[string]interface{})
	err := mapStructureDecoder(&D, &in, decoderConfig)
	if err != nil {
		return D, err
	}
	return D, nil
}

func flattenDatabaseInstanceDatastore(d dataStore) []map[string]interface{} {
	datastore := make([]map[string]interface{}, 1)
	datastore[0] = make(map[string]interface{})
	datastore[0]["type"] = d.Type
	datastore[0]["version"] = d.Version
	return datastore
}

func extractDatabaseNetworks(v []interface{}) ([]networkOpts, error) {
	Networks := make([]networkOpts, len(v))
	for i, network := range v {
		var N networkOpts
		err := mapstructure.Decode(network.(map[string]interface{}), &N)
		if err != nil {
			return nil, err
		}
		Networks[i] = N
	}
	return Networks, nil
}

func extractDatabaseAutoExpand(v []interface{}) (instanceAutoExpandOpts, error) {
	var A instanceAutoExpandOpts
	in := v[0].(map[string]interface{})
	err := mapstructure.Decode(in, &A)
	if err != nil {
		return A, err
	}
	return A, nil
}

func flattenDatabaseInstanceAutoExpand(autoExpandInt int, maxDiskSize int) []map[string]interface{} {
	autoExpand := make([]map[string]interface{}, 1)
	autoExpand[0] = make(map[string]interface{})
	if autoExpandInt != 0 {
		autoExpand[0]["autoexpand"] = true
	} else {
		autoExpand[0]["autoexpand"] = false
	}
	autoExpand[0]["max_disk_size"] = maxDiskSize
	return autoExpand
}

func extractDatabaseWalVolume(v []interface{}) (walVolumeOpts, error) {
	var W walVolumeOpts
	in := v[0].(map[string]interface{})
	err := mapstructure.Decode(in, &W)
	if err != nil {
		return W, err
	}
	return W, nil
}

func flattenDatabaseInstanceWalVolume(w walVolume) []map[string]interface{} {
	walvolume := make([]map[string]interface{}, 1)
	walvolume[0] = make(map[string]interface{})
	walvolume[0]["size"] = w.Size
	return walvolume
}

func extractDatabaseCapabilities(v []interface{}) ([]instanceCapabilityOpts, error) {
	capabilities := make([]instanceCapabilityOpts, len(v))
	for i, capability := range v {
		var C instanceCapabilityOpts
		err := mapstructure.Decode(capability.(map[string]interface{}), &C)
		if err != nil {
			return nil, err
		}
		capabilities[i] = C
	}
	return capabilities, nil
}

func flattenDatabaseInstanceCapabilities(c []instanceCapabilityOpts) []map[string]interface{} {
	capabilities := make([]map[string]interface{}, len(c))
	for i, capability := range c {
		capabilities[i] = make(map[string]interface{})
		capabilities[i]["name"] = capability.Name
		capabilities[i]["settings"] = capability.Params
	}
	return capabilities
}

func databaseInstanceStateRefreshFunc(client databaseClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		i, err := instanceGet(client, instanceID).extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); ok {
				return i, "DELETED", nil
			}
			return nil, "", err
		}

		if i.Status == "error" {
			return i, i.Status, fmt.Errorf("there was an error creating the database instance")
		}

		return i, i.Status, nil
	}
}

func getDBMSResource(client databaseClient, dbmsID string) (interface{}, error) {
	instanceResource, err := instanceGet(client, dbmsID).extract()
	if err == nil {
		return instanceResource, nil
	}
	clusterResource, err := dbClusterGet(client, dbmsID).extract()
	if err == nil {
		return clusterResource, nil
	}
	return nil, err
}
