package populator

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"graph-populator/parser"
)

// Graph handles Neo4j database operations for the knowledge graph.
type Graph struct {
	driver neo4j.DriverWithContext
	dbName string
}

// NewGraph creates a new Neo4j graph connection.
func NewGraph(uri, user, password, dbName string) (*Graph, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("create neo4j driver: %w", err)
	}
	return &Graph{driver: driver, dbName: dbName}, nil
}

// Close closes the Neo4j driver connection.
func (g *Graph) Close(ctx context.Context) error {
	return g.driver.Close(ctx)
}

// ClearAll removes all nodes and relationships in the database.
func (g *Graph) ClearAll(ctx context.Context) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.dbName})
	defer session.Close(ctx)
	_, err := session.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
	return err
}

// CreateNodes creates nodes of a given label from a list of property maps.
// Uses MERGE for idempotency so the populator can be re-run safely.
func (g *Graph) CreateNodes(ctx context.Context, label string, nodes []map[string]any) error {
	if len(nodes) == 0 {
		return nil
	}
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.dbName})
	defer session.Close(ctx)

	for _, props := range nodes {
		id, ok := props["id"]
		if !ok {
			continue
		}
		_, err := session.Run(ctx,
			fmt.Sprintf("MERGE (n:%s {id: $id}) SET n = $props", label),
			map[string]any{"id": id, "props": props},
		)
		if err != nil {
			return fmt.Errorf("create %s node %v: %w", label, id, err)
		}
	}
	return nil
}

// CreateRels creates relationships between existing nodes using MERGE.
func (g *Graph) CreateRels(ctx context.Context, rels []parser.RelDef) error {
	if len(rels) == 0 {
		return nil
	}
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.dbName})
	defer session.Close(ctx)

	for _, r := range rels {
		query := fmt.Sprintf(
			"MATCH (a {id: $fromId}) MATCH (b {id: $toId}) MERGE (a)-[:%s]->(b)",
			r.Type,
		)
		_, err := session.Run(ctx, query, map[string]any{
			"fromId": r.FromID,
			"toId":   r.ToID,
		})
		if err != nil {
			return fmt.Errorf("create rel %s (%s)->(%s): %w", r.Type, r.FromID, r.ToID, err)
		}
	}
	return nil
}
