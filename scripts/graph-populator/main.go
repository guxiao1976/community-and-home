package main

import (
	"context"
	"log"
	"os"

	"graph-populator/parser"
	"graph-populator/populator"
)

func main() {
	neo4jURI := getEnv("NEO4J_URI", "bolt://localhost:7687")
	neo4jUser := getEnv("NEO4J_USER", "neo4j")
	neo4jPass := getEnv("NEO4J_PASSWORD", "neo4j123456")
	neo4jDB := getEnv("NEO4J_DB", "neo4j")
	projectRoot := getEnv("PROJECT_ROOT", ".")

	ctx := context.Background()

	// Connect to Neo4j
	graph, err := populator.NewGraph(neo4jURI, neo4jUser, neo4jPass, neo4jDB)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer graph.Close(ctx)

	// Clear existing project data
	log.Println("Clearing existing graph data...")
	if err := graph.ClearAll(ctx); err != nil {
		log.Fatalf("Failed to clear graph: %v", err)
	}

	// Parse and populate Proto files
	log.Println("Parsing Proto files...")
	protoRes, err := parser.ParseProtos(projectRoot)
	if err != nil {
		log.Fatalf("Failed to parse proto files: %v", err)
	}
	if err := graph.CreateNodes(ctx, "ProtoMessage", protoRes.Nodes.Messages); err != nil {
		log.Printf("Error creating ProtoMessage nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "ProtoField", protoRes.Nodes.Fields); err != nil {
		log.Printf("Error creating ProtoField nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "ProtoRpc", protoRes.Nodes.Rpcs); err != nil {
		log.Printf("Error creating ProtoRpc nodes: %v", err)
	}
	if err := graph.CreateRels(ctx, protoRes.Rels); err != nil {
		log.Printf("Error creating proto relationships: %v", err)
	}
	log.Printf("  Proto: %d messages, %d fields, %d RPCs, %d relationships",
		len(protoRes.Nodes.Messages), len(protoRes.Nodes.Fields),
		len(protoRes.Nodes.Rpcs), len(protoRes.Rels))

	// Parse and populate Go sources
	log.Println("Parsing Go sources...")
	goRes, err := parser.ParseGoSources(projectRoot)
	if err != nil {
		log.Fatalf("Failed to parse Go sources: %v", err)
	}
	if err := graph.CreateNodes(ctx, "Service", goRes.Nodes.Services); err != nil {
		log.Printf("Error creating Service nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "GoStruct", goRes.Nodes.Structs); err != nil {
		log.Printf("Error creating GoStruct nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "GoField", goRes.Nodes.Fields); err != nil {
		log.Printf("Error creating GoField nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "ApiRoute", goRes.Nodes.Routes); err != nil {
		log.Printf("Error creating ApiRoute nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "DbTable", goRes.Nodes.Tables); err != nil {
		log.Printf("Error creating DbTable nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "DbColumn", goRes.Nodes.Columns); err != nil {
		log.Printf("Error creating DbColumn nodes: %v", err)
	}
	if err := graph.CreateRels(ctx, goRes.Rels); err != nil {
		log.Printf("Error creating Go relationships: %v", err)
	}
	log.Printf("  Go: %d services, %d structs, %d fields, %d routes, %d tables, %d columns, %d relationships",
		len(goRes.Nodes.Services), len(goRes.Nodes.Structs),
		len(goRes.Nodes.Fields), len(goRes.Nodes.Routes),
		len(goRes.Nodes.Tables), len(goRes.Nodes.Columns), len(goRes.Rels))

	// Parse and populate TypeScript sources
	log.Println("Parsing TypeScript sources...")
	tsRes, err := parser.ParseTypeScript(projectRoot)
	if err != nil {
		log.Fatalf("Failed to parse TypeScript sources: %v", err)
	}
	if err := graph.CreateNodes(ctx, "TsInterface", tsRes.Nodes.Interfaces); err != nil {
		log.Printf("Error creating TsInterface nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "TsField", tsRes.Nodes.Fields); err != nil {
		log.Printf("Error creating TsField nodes: %v", err)
	}
	if err := graph.CreateNodes(ctx, "ApiCall", tsRes.Nodes.ApiCalls); err != nil {
		log.Printf("Error creating ApiCall nodes: %v", err)
	}
	if err := graph.CreateRels(ctx, tsRes.Rels); err != nil {
		log.Printf("Error creating TypeScript relationships: %v", err)
	}
	log.Printf("  TypeScript: %d interfaces, %d fields, %d API calls, %d relationships",
		len(tsRes.Nodes.Interfaces), len(tsRes.Nodes.Fields),
		len(tsRes.Nodes.ApiCalls), len(tsRes.Rels))

	// Parse and populate infrastructure config
	log.Println("Parsing infrastructure config...")
	infraRes, err := parser.ParseInfra(projectRoot)
	if err != nil {
		log.Fatalf("Failed to parse infra config: %v", err)
	}
	if err := graph.CreateNodes(ctx, "Service", infraRes.Nodes.Services); err != nil {
		log.Printf("Error creating infra Service nodes: %v", err)
	}
	if err := graph.CreateRels(ctx, infraRes.Rels); err != nil {
		log.Printf("Error creating infra relationships: %v", err)
	}
	log.Printf("  Infra: %d services, %d relationships",
		len(infraRes.Nodes.Services), len(infraRes.Rels))

	// Cross-source alignment
	log.Println("Aligning cross-source references...")
	alignRels := parser.AlignCrossSource(protoRes, goRes, tsRes)
	if err := graph.CreateRels(ctx, alignRels); err != nil {
		log.Printf("Error creating cross-source relationships: %v", err)
	}
	log.Printf("  Cross-source: %d relationships", len(alignRels))

	log.Println("Graph population complete!")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
