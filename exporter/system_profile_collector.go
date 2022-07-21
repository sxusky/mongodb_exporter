package exporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type systemProfileCollector struct {
	ctx  context.Context
	base *baseCollector

	compatibleMode bool
	topologyInfo   labelsGetter
}

// newProfileCollector creates a collector for system.profile.
func newSystemProfileCollector(ctx context.Context, client *mongo.Client, logger *logrus.Logger, compatible bool, topology labelsGetter) *systemProfileCollector {
	return &systemProfileCollector{
		ctx:  ctx,
		base: newBaseCollector(client, logger),

		compatibleMode: compatible,
		topologyInfo:   topology,
	}
}

func (d *systemProfileCollector) Describe(ch chan<- *prometheus.Desc) {
	d.base.Describe(d.ctx, ch, d.collect)
}

func (d *systemProfileCollector) Collect(ch chan<- prometheus.Metric) {
	d.base.Collect(ch)
}

func (d *systemProfileCollector) collect(ch chan<- prometheus.Metric) {
	var m bson.M

	logger := d.base.logger
	client := d.base.client

	cmd := bson.D{{Key: "find", Value: "system.profile"}}
	res := client.Database("test").RunCommand(d.ctx, cmd)

	debugResult(logger, res)
	if err := res.Decode(&m); err != nil {
		logger.Errorf("cannot find system.profile: %s", err)
	}

	if m == nil || m["cursor"] == nil {
		logger.Error("cannot find system.profile: response is empty")
	}

	logger.Debug("system.profile result")
	debugResult(logger, m)

	metrics := makeMetrics("", m, d.topologyInfo.baseLabels(), d.compatibleMode)
	metrics = append(metrics, locksMetrics(m)...)

	if d.compatibleMode {
		metrics = append(metrics, specialMetrics(d.ctx, client, m, logger)...)

		if cem, err := cacheEvictedTotalMetric(m); err == nil {
			metrics = append(metrics, cem)
		}

		nodeType, err := getNodeType(d.ctx, client)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"component": "systemProfileCollector",
			}).Errorf("Cannot get node type to check if this is a mongos: %s", err)
		} else if nodeType == typeMongos {
			metrics = append(metrics, mongosMetrics(d.ctx, client, logger)...)
		}
	}

	for _, metric := range metrics {
		ch <- metric
	}
}

// check interface.
var _ prometheus.Collector = (*systemProfileCollector)(nil)
