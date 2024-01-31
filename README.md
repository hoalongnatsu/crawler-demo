# Crawler Demo for 200LAB DevOps Course

In this course, we will set up the CI/CD and infrastructure for the crawler systems below on Kubernetes.

## Overview System

Crawler System: crawler posts + redis stream + consumer

![picture](./images/crawler.png)

Change Data Capture System: Postgres + Kafka Connect + Elasticsearch

![picture](./images/cdc.png)

API Service: Rust + Postgres + Elasticsearch

![picture](./images/api.png)