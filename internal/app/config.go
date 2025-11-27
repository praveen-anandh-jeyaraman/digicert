package app

import (
    "errors"
    "os"
)

type Config struct {
    DatabaseURL string
    Port        string

    // AWS CloudWatch
    Region              string
    CloudWatchLogGroup  string
    CloudWatchLogStream string
    EnableCloudWatch    bool
}

func LoadConfigFromEnv() (*Config, error) {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        return nil, errors.New("DATABASE_URL required")
    }
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    return &Config{
        DatabaseURL: dsn,
        Port:        port,

        // AWS CloudWatch config
        Region:              getEnv("AWS_REGION", "us-east-1"),
        CloudWatchLogGroup:  getEnv("CW_LOG_GROUP", "/aws/ec2/library-api"),
        CloudWatchLogStream: getEnv("CW_LOG_STREAM", "library-api"),
        EnableCloudWatch:    getEnv("ENABLE_CLOUDWATCH", "true") == "true",
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}