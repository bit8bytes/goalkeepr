package main

type contextKey string

const (
	TraceIdKey contextKey = "requestTraceIdKey"
	ConnIdKey  contextKey = "requestConnIdKey"
	UserIdKey  contextKey = "userID"
)
