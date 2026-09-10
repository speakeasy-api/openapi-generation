package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_PythonChanges(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Python SDK API
  version: 0.1.0
paths:
  /items:
    get:
      summary: List items
      operationId: listItems
      responses:
        '200':
          description: A list of items
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Item'
components:
  schemas:
    Item:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        price:
          type: number
`

	newSpec := `openapi: 3.0.0
info:
  title: Python SDK API
  version: 0.2.0
paths:
  /items:
    get:
      summary: List items with pagination
      operationId: listItems
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 10
      responses:
        '200':
          description: A paginated list of items
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Item'
  /items/{itemId}:
    delete:
      summary: Delete an item
      operationId: deleteItem
      parameters:
        - name: itemId
          in: path
          required: true
          schema:
            type: string
      responses:
        '204':
          description: Item deleted successfully
components:
  schemas:
    Item:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        price:
          type: number
        category:
          type: string
          enum: ['electronics', 'clothing', 'food', 'other']
`

	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: oldSpec,
		newSpec: newSpec,
		lang:    "python",
	})

	diff, err := Changes(t.Context(), oldConfig, newConfig)
	if err != nil {
		t.Fatalf("failed to get changes. error: %s", err.Error())
	}
	// Snapshot both compact and full modes

	// Use go-snaps to snapshot the markdown output
	// Include specs and output for easier review
	snapshot := formatTestSnapshotWithDiff(t.Name(), oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
