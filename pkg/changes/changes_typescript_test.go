package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_TypeScriptChanges(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: TypeScript SDK API
  version: 1.0.0
paths:
  /products:
    get:
      summary: Get products
      operationId: getProducts
      parameters:
        - name: category
          in: query
          schema:
            type: string
      responses:
        '200':
          description: List of products
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
                    name:
                      type: string
                    price:
                      type: number
                      format: double
`

	newSpec := `openapi: 3.0.0
info:
  title: TypeScript SDK API
  version: 1.1.0
paths:
  /products:
    get:
      summary: Get products with filters
      operationId: getProducts
      parameters:
        - name: request
          in: query
          required: true
          style: deepObject
          explode: true
          schema:
            type: object
            properties:
              category:
                type: string
              minPrice:
                type: number
              maxPrice:
                type: number
              inStock:
                type: boolean
      responses:
        '200':
          description: Filtered list of products
          content:
            application/json:
              schema:
                type: object
                properties:
                  products:
                    type: array
                    items:
                      type: object
                      properties:
                        id:
                          type: string
                        name:
                          type: string
                        price:
                          type: number
                          format: double
                        stock:
                          type: integer
                        tags:
                          type: array
                          items:
                            type: string
                  metadata:
                    type: object
                    properties:
                      total:
                        type: integer
                      filtered:
                        type: integer
    post:
      summary: Create a new product
      operationId: createProduct
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - price
              properties:
                name:
                  type: string
                price:
                  type: number
                  format: double
                description:
                  type: string
      responses:
        '201':
          description: Product created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
                  price:
                    type: number
                  createdAt:
                    type: string
                    format: date-time
`

	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: oldSpec,
		newSpec: newSpec,
		lang:    "typescript",
	})

	diff, err := Changes(t.Context(), oldConfig, newConfig)
	if err != nil {
		t.Fatalf("failed to get changes. error: %s", err.Error())
	}
	// Snapshot both compact and full modes

	// Include specs and output for easier review
	snapshot := formatTestSnapshotWithDiff(t.Name(), oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
