/**
 * Copyright 2018 Confluent Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kafka

import (
	"context"
	"fmt"
	"time"
	"unsafe"
)

/*
#include <librdkafka/rdkafka.h>
#include <stdlib.h>

static const rd_kafka_topic_result_t *
topic_result_by_idx (const rd_kafka_topic_result_t **topics, size_t cnt, size_t idx) {
    if (idx >= cnt)
      return NULL;
    return topics[idx];
}
*/
import "C"

// AdminClient is derived from an existing Producer or Consumer
type AdminClient struct {
	handle    *handle
	isDerived bool // Derived from existing client handle
}

// AdminOptions FIXME
type AdminOptions struct {
	OperationTimeout time.Duration
	RequestTimeout   time.Duration
	ValidateOnly     bool
}

func durationToMilliseconds(t time.Duration) int {
	if t > 0 {
		return (int)(t.Seconds() * 1000.0)
	}
	return (int)(t)
}

// adminOptionsToC converts Golang AdminOptions to C AdminOptions
// forAPI is used to limit what options may be set.
func (a *AdminClient) adminOptionsToC(forAPI string, options *AdminOptions) (cOptions *C.rd_kafka_AdminOptions_t, err error) {
	if options == nil {
		return nil, nil
	}

	// FIXME: enforce forAPI
	cOptions = C.rd_kafka_AdminOptions_new(a.handle.rk)

	cErrstrSize := C.size_t(512)
	cErrstr := (*C.char)(C.malloc(cErrstrSize))
	defer C.free(unsafe.Pointer(cErrstr))

	if options.OperationTimeout != 0 {
		cErr := C.rd_kafka_AdminOptions_set_operation_timeout(
			cOptions, C.int(durationToMilliseconds(options.OperationTimeout)),
			cErrstr, cErrstrSize)
		if cErr != 0 {
			C.rd_kafka_AdminOptions_destroy(cOptions)
			return nil, newCErrorFromString(cErr,
				fmt.Sprintf("Failed to set operation timeout: %s", C.GoString(cErrstr)))

		}
	}

	if options.RequestTimeout != 0 {
		cErr := C.rd_kafka_AdminOptions_set_request_timeout(
			cOptions, C.int(durationToMilliseconds(options.RequestTimeout)),
			cErrstr, cErrstrSize)
		if cErr != 0 {
			C.rd_kafka_AdminOptions_destroy(cOptions)
			return nil, newCErrorFromString(cErr,
				fmt.Sprintf("Failed to set request timeout: %s", C.GoString(cErrstr)))
		}
	}

	cErr := C.rd_kafka_AdminOptions_set_validate_only(
		cOptions, bool2cint(options.ValidateOnly),
		cErrstr, cErrstrSize)
	if cErr != 0 {
		C.rd_kafka_AdminOptions_destroy(cOptions)
		return nil, newCErrorFromString(cErr,
			fmt.Sprintf("Failed to set validate only: %s", C.GoString(cErrstr)))
	}

	return cOptions, nil
}

// TopicResult FIXME
type TopicResult struct {
	// Topic name of result.
	Topic string
	// Error, if any, of result.
	Error Error
}

// String provides a human-readable representation of a TopicResult
func (t TopicResult) String() string {
	if t.Error.code == 0 {
		return t.Topic
	}
	return fmt.Sprintf("%s (%s)", t.Topic, t.Error.str)
}

// NewTopic FIXME
type NewTopic struct {
	// Topic name to create.
	Topic string
	// Number of partitions in topic.
	NumPartitions int
	// Default replication factor for the topic's partitions, or zero
	// if an explicit ReplicaAssignment is set.
	ReplicationFactor int
	// (Optional) Explicit replica assignment. The outer array is
	// indexed by the partition number, while the inner per-partition array
	// contains the replica broker ids. The first broker in each
	// broker id list will be the preferred replica.
	ReplicaAssignment [][]int
	Config            map[string]string
}

// NewPartitions FIXME
type NewPartitions struct {
	// Topic to create more partitions for.
	Topic string
	// New partition count for topic, must be higher than current partition count.
	NewTotalCount int
	// (Optional) Explicit replica assignment. The outer array is
	// indexed by the new partition index (i.e., 0 for the first added
	// partition), while the inner per-partition array
	// contains the replica broker ids. The first broker in each
	// broker id list will be the preferred replica.
	ReplicaAssignment [][]int
}

// ResourceType FIXME
type ResourceType int

const (
	// ResourceUnknown - Unknown
	ResourceUnknown = ResourceType(C.RD_KAFKA_RESOURCE_UNKNOWN)
	// ResourceAny - match any resource type (DescribeConfigs)
	ResourceAny = ResourceType(C.RD_KAFKA_RESOURCE_ANY)
	// ResourceTopic - Topic
	ResourceTopic = ResourceType(C.RD_KAFKA_RESOURCE_TOPIC)
	// ResourceGroup - Group
	ResourceGroup = ResourceType(C.RD_KAFKA_RESOURCE_GROUP)
	// ResourceBroker - Broker
	ResourceBroker = ResourceType(C.RD_KAFKA_RESOURCE_BROKER)
)

// String returns the human-readable representation of a ResourceType
func (t ResourceType) String() string {
	return C.GoString(C.rd_kafka_ResourceType_name(C.rd_kafka_ResourceType_t(t)))
}

// ConfigSource FIXME
type ConfigSource int

const (
	// ConfigSourceUnknown is the default value
	ConfigSourceUnknown = ConfigSource(C.RD_KAFKA_CONFIG_SOURCE_UNKNOWN_CONFIG)
	// ConfigSourceDynamicTopic is dynamic topic config that is configured for a specific topic
	ConfigSourceDynamicTopic = ConfigSource(C.RD_KAFKA_CONFIG_SOURCE_DYNAMIC_TOPIC_CONFIG)
	// ConfigSourceDynamicBroker is dynamic broker config that is configured for a specific broker
	ConfigSourceDynamicBroker = ConfigSource(C.RD_KAFKA_CONFIG_SOURCE_DYNAMIC_BROKER_CONFIG)
	// ConfigSourceDynamicDefaultBroker is dynamic broker config that is configured as default for all brokers in the cluster
	ConfigSourceDynamicDefaultBroker = ConfigSource(C.RD_KAFKA_CONFIG_SOURCE_DYNAMIC_DEFAULT_BROKER_CONFIG)
	// ConfigSourceStaticBroker is static broker config provided as broker properties at startup (e.g. from server.properties file)
	ConfigSourceStaticBroker = ConfigSource(C.RD_KAFKA_CONFIG_SOURCE_STATIC_BROKER_CONFIG)
	// ConfigSourceDefault is built-in default configuration for configs that have a default value
	ConfigSourceDefault = ConfigSource(C.RD_KAFKA_CONFIG_SOURCE_DEFAULT_CONFIG)
)

// String returns the human-readable representation of a ConfigSource type
func (t ConfigSource) String() string {
	return C.GoString(C.rd_kafka_ConfigSource_name(C.rd_kafka_ConfigSource_t(t)))
}

// ConfigResource FIXME
type ConfigResource struct {
	// Type of resource to set.
	Type ResourceType
	// Name of resource to set.
	Name string
	// Config entries to set.
	Config map[string]string
}

// ConfigEntry FIXME
type ConfigEntry struct {
	// Name of configuration entry, e.g., topic configuration property name.
	Name string
	// Value of configuration entry.
	Value string
	// Source indicates the configuration source.
	Source ConfigSource
	// IsReadOnly indicates whether the configuration entry can be altered.
	IsReadOnly bool
	// IsSensitive indicates whether the configuration entry contains sensitive information, in which case the value will be unset.
	IsSensitive bool
	// IsSynonym indicates whether the configuration entry is a synonym for another configuration property.
	IsSynonym bool
	// Synonyms contains a map of configuration entries that are synonyms to this configuration entry.
	Synonyms map[string]ConfigEntry
}

// ConfigResourceResult FIXME
type ConfigResourceResult struct {
	// Type of returned result resource.
	Type ResourceType
	// Name of returned result resource.
	Name string
	// Error, if any, of returned result resource.
	Error Error
	// Config entries, if any, of returned result resource.
	Config map[string]ConfigEntry
}

// waitResult waits for a result event on cQueue or the ctx to be cancelled, whichever happens
// first.
// The returned result event is checked for errors its error is returned if set.
func (a *AdminClient) waitResult(ctx context.Context, cQueue *C.rd_kafka_queue_t, cEventType C.rd_kafka_event_type_t) (rkev *C.rd_kafka_event_t, err error) {

	resultChan := make(chan *C.rd_kafka_event_t)
	closeChan := make(chan bool) // never written to, just closed

	go func() {
		for {
			select {
			case _, ok := <-closeChan:
				if !ok {
					// Context cancelled/timed out
					close(resultChan)
					return
				}

			default:
				// Wait for result event for at most 100ms
				rkev := C.rd_kafka_queue_poll(cQueue, 100)
				if rkev != nil {
					resultChan <- rkev
					close(resultChan)
					return
				}
			}
		}
	}()

	defer close(closeChan)

	select {
	case rkev = <-resultChan:
		// Result type check
		if cEventType != C.rd_kafka_event_type(rkev) {
			err = newErrorFromString(ErrInvalidType,
				fmt.Sprintf("Expected %d result event, not %d", (int)(cEventType), (int)(C.rd_kafka_event_type(rkev))))
			C.rd_kafka_event_destroy(rkev)
			return nil, err
		}

		// Generic error handling
		cErr := C.rd_kafka_event_error(rkev)
		if cErr != 0 {
			err = newErrorFromCString(cErr, C.rd_kafka_event_error_string(rkev))
			C.rd_kafka_event_destroy(rkev)
			return nil, err
		}
		return rkev, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (a *AdminClient) cToTopicResults(cTopicRes **C.rd_kafka_topic_result_t, cCnt C.size_t) (result []TopicResult, err error) {

	result = make([]TopicResult, int(cCnt))

	for i := 0; i < int(cCnt); i++ {
		cTopic := C.topic_result_by_idx(cTopicRes, cCnt, C.size_t(i))
		result[i].Topic = C.GoString(C.rd_kafka_topic_result_name(cTopic))
		result[i].Error = newErrorFromCString(
			C.rd_kafka_topic_result_error(cTopic),
			C.rd_kafka_topic_result_error_string(cTopic))
	}

	return result, nil
}

// CreateTopics FIXME
func (a *AdminClient) CreateTopics(ctx context.Context, topics []NewTopic, options *AdminOptions) (result []TopicResult, err error) {
	cTopics := make([]*C.rd_kafka_NewTopic_t, len(topics))

	cErrstrSize := C.size_t(512)
	cErrstr := (*C.char)(C.malloc(cErrstrSize))
	defer C.free(unsafe.Pointer(cErrstr))

	for i, topic := range topics {

		var cReplicationFactor C.int
		if topic.ReplicationFactor == 0 {
			cReplicationFactor = -1
		} else {
			cReplicationFactor = C.int(topic.ReplicationFactor)
		}
		if topic.ReplicaAssignment != nil {
			if cReplicationFactor != -1 {
				return nil, newErrorFromString(ErrInvalidArg,
					"NewTopic.ReplicationFactor and NewTopic.ReplicaAssignment are mutually exclusive")
			}

			if len(topic.ReplicaAssignment) != topic.NumPartitions {
				return nil, newErrorFromString(ErrInvalidArg,
					"NewTopic.ReplicaAssignment must contain exactly NewTopic.NumPartitions partitions")
			}

		} else if cReplicationFactor == -1 {
			return nil, newErrorFromString(ErrInvalidArg,
				"NewTopic.ReplicationFactor or NewTopic.ReplicaAssignment must be specified")
		}

		cTopics[i] = C.rd_kafka_NewTopic_new(
			C.CString(topic.Topic),
			C.int(topic.NumPartitions),
			cReplicationFactor)
		if cTopics[i] == nil {
			return nil, newErrorFromString(ErrInvalidArg,
				fmt.Sprintf("Invalid arguments for topic %s", topic.Topic))
		}

		defer C.rd_kafka_NewTopic_destroy(cTopics[i])

		for p, replicas := range topic.ReplicaAssignment {
			cReplicas := make([]C.int32_t, len(replicas))
			for ri, replica := range replicas {
				cReplicas[ri] = C.int32_t(replica)
			}
			cErr := C.rd_kafka_NewTopic_set_replica_assignment(
				cTopics[i], C.int32_t(p),
				(*C.int32_t)(&cReplicas[0]), C.size_t(len(cReplicas)),
				cErrstr, cErrstrSize)
			if cErr != 0 {
				return nil, newCErrorFromString(cErr,
					fmt.Sprintf("Failed to set replica assignment for topic %s partition %d: %s", topic.Topic, p, C.GoString(cErrstr)))
			}
		}

		for k, v := range topic.Config {
			cErr := C.rd_kafka_NewTopic_add_config(cTopics[i], C.CString(k), C.CString(v))
			if cErr != 0 {
				return nil, newCErrorFromString(cErr,
					fmt.Sprintf("Failed to add config \"%s\" for topic %s", k, topic.Topic))
			}
		}
	}

	// Convert Go AdminOptions (if any) to C AdminOptions
	cOptions, err := a.adminOptionsToC("CreateTopics", options)
	if err != nil {
		return nil, err
	}

	if cOptions != nil {
		defer C.rd_kafka_AdminOptions_destroy(cOptions)
	}

	// Create temporary queue for async operation
	cQueue := C.rd_kafka_queue_new(a.handle.rk)
	defer C.rd_kafka_queue_destroy(cQueue)

	// Asynchronous call
	C.rd_kafka_CreateTopics(
		a.handle.rk,
		(**C.rd_kafka_NewTopic_t)(&cTopics[0]),
		C.size_t(len(cTopics)),
		cOptions,
		cQueue)

	// Wait for result, error or context timeout
	rkev, err := a.waitResult(ctx, cQueue, C.RD_KAFKA_EVENT_CREATETOPICS_RESULT)
	if err != nil {
		return nil, err
	}
	defer C.rd_kafka_event_destroy(rkev)

	cRes := C.rd_kafka_event_CreateTopics_result(rkev)

	// Convert result from C to Go
	var cCnt C.size_t
	cTopicRes := C.rd_kafka_CreateTopics_result_topics(cRes, &cCnt)

	return a.cToTopicResults(cTopicRes, cCnt)
}

// DeleteTopics FIXME
func DeleteTopics(ctx context.Context, topics []string, options AdminOptions) (result []TopicResult, err error) {
	return nil, nil

}

// CreatePartitions FIXME
func CreatePartitions(ctx context.Context, partitions []NewPartitions, options AdminOptions) (result []TopicResult, err error) {
	return nil, nil
}

// AlterConfigs FIXME
func AlterConfigs(ctx context.Context, resources []ConfigResource, options AdminOptions) (result []ConfigResourceResult, err error) {
	return nil, nil
}

// DescribeConfigs FIXME
func DescribeConfigs(ctx context.Context, resources []ConfigResource, options AdminOptions) (result []ConfigResourceResult, err error) {
	return nil, nil
}

// String returns a human readable name for an AdminClient instance
func (a *AdminClient) String() string {
	return fmt.Sprintf("admin-%s", a.handle.String())
}

// Close an AdminClient instance.
func (a *AdminClient) Close() {
	if a.isDerived {
		// Derived AdminClient needs no cleanup.
		a.handle = &handle{}
		return
	}

	a.handle.cleanup()

	C.rd_kafka_destroy(a.handle.rk)
}

// NewAdminClient creats a new AdminClient instance with a new underlying client instance
func NewAdminClient(conf *ConfigMap) (*AdminClient, error) {

	err := versionCheck()
	if err != nil {
		return nil, err
	}

	a := &AdminClient{}
	a.handle = &handle{}

	// Convert ConfigMap to librdkafka conf_t
	cConf, err := conf.convert()
	if err != nil {
		return nil, err
	}

	cErrstr := (*C.char)(C.malloc(C.size_t(256)))
	defer C.free(unsafe.Pointer(cErrstr))

	// Create librdkafka producer instance. The Producer is somewhat cheaper than
	// the consumer, but any instance type can be used for Admin APIs.
	a.handle.rk = C.rd_kafka_new(C.RD_KAFKA_PRODUCER, cConf, cErrstr, 256)
	if a.handle.rk == nil {
		return nil, newErrorFromCString(C.RD_KAFKA_RESP_ERR__INVALID_ARG, cErrstr)
	}

	a.isDerived = false
	a.handle.setup()

	return a, nil
}

// NewAdminClientFromProducer derives a new AdminClient from an existing Producer instance.
// The AdminClient will use the same configuration and connections as the parent instance.
func NewAdminClientFromProducer(p *Producer) (a *AdminClient, err error) {
	if p.handle.rk == nil {
		return nil, newErrorFromString(ErrInvalidArg, "Can't derive AdminClient from closed producer")
	}

	a = &AdminClient{}
	a.handle = &p.handle
	a.isDerived = true
	return a, nil
}

// NewAdminClientFromConsumer derives a new AdminClient from an existing Consumer instance.
// The AdminClient will use the same configuration and connections as the parent instance.
func NewAdminClientFromConsumer(c *Consumer) (a *AdminClient, err error) {
	if c.handle.rk == nil {
		return nil, newErrorFromString(ErrInvalidArg, "Can't derive AdminClient from closed consumer")
	}

	a = &AdminClient{}
	a.handle = &c.handle
	a.isDerived = true
	return a, nil
}
