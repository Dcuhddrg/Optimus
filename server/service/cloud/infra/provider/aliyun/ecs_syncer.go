package aliyun

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"optimus/server/service/cloud/usecase"
)

// ECSSyncer implements AssetSyncer for Aliyun ECS instances.
// Credential map must contain: access_key_id, access_key_secret.
// Optional: region (defaults to "cn-hangzhou").
type ECSSyncer struct{}

func NewECSSyncer() *ECSSyncer { return &ECSSyncer{} }

func (s *ECSSyncer) Sync(ctx context.Context, account usecase.SyncAccount) ([]usecase.RawAsset, error) {
	ak := account.Credential["access_key_id"]
	sk := account.Credential["access_key_secret"]
	region := account.Credential["region"]
	if region == "" {
		region = "cn-hangzhou"
	}

	client, err := ecs.NewClientWithAccessKey(region, ak, sk)
	if err != nil {
		return nil, fmt.Errorf("aliyun ecs client: %w", err)
	}

	var assets []usecase.RawAsset
	pageNumber := 1
	for {
		req := ecs.CreateDescribeInstancesRequest()
		req.PageSize = requests.NewInteger(100)
		req.PageNumber = requests.NewInteger(pageNumber)

		resp, err := client.DescribeInstances(req)
		if err != nil {
			return nil, fmt.Errorf("aliyun DescribeInstances: %w", err)
		}

		for _, inst := range resp.Instances.Instance {
			configJSON, _ := json.Marshal(map[string]any{
				"instanceType":       inst.InstanceType,
				"cpu":                inst.Cpu,
				"memory":             inst.Memory,
				"osType":             inst.OSType,
				"instanceChargeType": inst.InstanceChargeType,
			})
			tagsMap := make(map[string]string, len(inst.Tags.Tag))
			for _, tag := range inst.Tags.Tag {
				tagsMap[tag.TagKey] = tag.TagValue
			}
			tagsJSON, _ := json.Marshal(tagsMap)

			assets = append(assets, usecase.RawAsset{
				ResourceType: "ecs_instance",
				ResourceID:   inst.InstanceId,
				Name:         inst.InstanceName,
				Region:       region,
				Status:       inst.Status,
				ConfigJSON:   string(configJSON),
				TagsJSON:     string(tagsJSON),
			})
		}

		if pageNumber*100 >= resp.TotalCount {
			break
		}
		pageNumber++
	}
	return assets, nil
}
