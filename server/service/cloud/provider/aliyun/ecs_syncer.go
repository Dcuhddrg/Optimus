package aliyun

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	cloudService "optimus/server/service/cloud"
)

// ECSSyncer implements AssetSyncer for Aliyun ECS instances.
// Credential map must contain: access_key_id, access_key_secret.
// Optional: region (defaults to "cn-hangzhou").
type ECSSyncer struct{}

func NewECSSyncer() *ECSSyncer { return &ECSSyncer{} }

func (s *ECSSyncer) Sync(ctx context.Context, account cloudService.SyncAccount) ([]cloudService.RawAsset, error) {
	ak := account.Credential["access_key_id"]
	sk := account.Credential["access_key_secret"]
	region := account.Credential["region"]
	if region == "" {
		region = "cn-hangzhou"
	}

	if ak == "" || sk == "" {
		return nil, fmt.Errorf("aliyun ecs: missing access_key_id or access_key_secret in credentials")
	}

	client, err := ecs.NewClientWithAccessKey(region, ak, sk)
	if err != nil {
		return nil, fmt.Errorf("aliyun ecs client: %w", err)
	}

	var assets []cloudService.RawAsset
	pageNumber := 1
	for {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("aliyun ecs sync cancelled: %w", err)
		}
		req := ecs.CreateDescribeInstancesRequest()
		req.PageSize = requests.NewInteger(100)
		req.PageNumber = requests.NewInteger(pageNumber)

		resp, err := client.DescribeInstances(req)
		if err != nil {
			return nil, fmt.Errorf("aliyun DescribeInstances: %w", err)
		}

		for _, inst := range resp.Instances.Instance {
			configJSON, err := json.Marshal(map[string]any{
				"instanceType":       inst.InstanceType,
				"cpu":                inst.Cpu,
				"memory":             inst.Memory,
				"osType":             inst.OSType,
				"instanceChargeType": inst.InstanceChargeType,
			})
			if err != nil {
				return nil, fmt.Errorf("aliyun ecs marshal config for %s: %w", inst.InstanceId, err)
			}
			tagsMap := make(map[string]string, len(inst.Tags.Tag))
			for _, tag := range inst.Tags.Tag {
				tagsMap[tag.TagKey] = tag.TagValue
			}
			tagsJSON, err := json.Marshal(tagsMap)
			if err != nil {
				return nil, fmt.Errorf("aliyun ecs marshal tags for %s: %w", inst.InstanceId, err)
			}

			assets = append(assets, cloudService.RawAsset{
				ResourceType: "ecs_instance",
				ResourceID:   inst.InstanceId,
				Name:         inst.InstanceName,
				Region:       region,
				Status:       inst.Status,
				ConfigJSON:   string(configJSON),
				TagsJSON:     string(tagsJSON),
			})
		}

		if len(resp.Instances.Instance) < 100 {
			break
		}
		pageNumber++
	}
	return assets, nil
}
