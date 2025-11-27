package logger

import (
    "context"
    "fmt"
    "log"
    "os"
    "sync"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type CloudWatchLogger struct {
    client     *cloudwatch.Client
    logGroup   string
    logStream  string
    stdLogger  *log.Logger
    mu         sync.Mutex
    logBuffer  []string
    isEnabled  bool
}

func NewCloudWatchLogger(logGroup, logStream string, enabled bool) (*CloudWatchLogger, error) {
    if !enabled {
        return &CloudWatchLogger{
            stdLogger: log.New(os.Stdout, "", log.LstdFlags),
            isEnabled: false,
        }, nil
    }

    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    client := cloudwatch.NewFromConfig(cfg)

    cwLogger := &CloudWatchLogger{
        client:    client,
        logGroup:  logGroup,
        logStream: logStream,
        stdLogger: log.New(os.Stdout, "", log.LstdFlags),
        logBuffer: make([]string, 0, 100),
        isEnabled: true,
    }

    log.SetOutput(cwLogger)

    return cwLogger, nil
}

func (l *CloudWatchLogger) Write(p []byte) (n int, err error) {
    message := string(p)

    // Always write to stdout
    l.stdLogger.Print(message)

    if l.isEnabled {
        l.mu.Lock()
        l.logBuffer = append(l.logBuffer, message)
        l.mu.Unlock()
    }

    return len(p), nil
}

// Flush sends buffered logs to CloudWatch
func (l *CloudWatchLogger) Flush(ctx context.Context) error {
    if !l.isEnabled {
        return nil
    }

    l.mu.Lock()
    if len(l.logBuffer) == 0 {
        l.mu.Unlock()
        return nil
    }

    logs := l.logBuffer
    l.logBuffer = make([]string, 0, 100)
    l.mu.Unlock()

    // Send metric to CloudWatch about log count
    _, err := l.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
        Namespace: aws.String("LibraryAPI"),
        MetricData: []types.MetricDatum{
            {
                MetricName: aws.String("LogLines"),
                Value:      aws.Float64(float64(len(logs))),
                Unit:       types.StandardUnitCount,
            },
        },
    })

    return err
}

// PutMetric sends a custom metric to CloudWatch
func (l *CloudWatchLogger) PutMetric(ctx context.Context, metricName string, value float64, unit string) error {
    if !l.isEnabled {
        return nil
    }

    _, err := l.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
        Namespace: aws.String("LibraryAPI"),
        MetricData: []types.MetricDatum{
            {
                MetricName: aws.String(metricName),
                Value:      aws.Float64(value),
                Unit:       types.StandardUnit(unit),
            },
        },
    })
    return err
}

// Close closes the CloudWatch client
func (l *CloudWatchLogger) Close() error {
    if !l.isEnabled {
        return nil
    }
    return nil
}

var globalLogger *CloudWatchLogger

// Initialize sets up the global logger
func Initialize(logGroup, logStream string, enabled bool) error {
    cwLogger, err := NewCloudWatchLogger(logGroup, logStream, enabled)
    if err != nil {
        return err
    }
    globalLogger = cwLogger
    return nil
}

// GetLogger returns the global logger
func GetLogger() *CloudWatchLogger {
    if globalLogger == nil {
        globalLogger = &CloudWatchLogger{
            stdLogger: log.New(os.Stdout, "", log.LstdFlags),
            isEnabled: false,
        }
    }
    return globalLogger
}