package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_CompleteSpecWithAllChangeTypes(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Complete Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      parameters:
        - name: X-Request-ID
          in: header
          schema:
            type: string
        - name: page
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: integer
                    name:
                      type: string
                    profile:
                      type: object
                      properties:
                        personal:
                          type: object
                          properties:
                            contact:
                              type: object
                              properties:
                                email:
                                  type: string
                                phones:
                                  type: array
                                  items:
                                    type: string
                        preferences:
                          type: object
                          additionalProperties:
                            type: string
                    status:
                      oneOf:
                        - type: string
                        - type: integer
    post:
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required:
                - name
              properties:
                name:
                  type: string
                email:
                  type: string
                  nullable: true
                profile:
                  type: object
                  properties:
                    personal:
                      type: object
                      properties:
                        contact:
                          type: object
                          properties:
                            email:
                              type: string
                            phones:
                              type: array
                              items:
                                type: string
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
  /users/{id}:
    get:
      operationId: getUserById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
    delete:
      operationId: deleteUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '204':
          description: No Content`

	newSpec := `openapi: 3.0.0
info:
  title: Complete Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      parameters:
        - name: X-Request-ID
          in: header
          schema:
            type: string
        - name: X-API-Key
          in: header
          required: true
          schema:
            type: string
        - name: page
          in: query
          schema:
            type: integer
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  required:
                    - id
                    - name
                  properties:
                    id:
                      type: integer
                    name:
                      type: string
                    email:
                      type: string
                    profile:
                      type: object
                      properties:
                        personal:
                          type: object
                          properties:
                            contact:
                              type: object
                              properties:
                                email:
                                  type: string
                                phones:
                                  type: array
                                  items:
                                    type: integer
                                address:
                                  type: object
                                  properties:
                                    street:
                                      type: string
                                    city:
                                      type: string
                        preferences:
                          type: object
                          additionalProperties:
                            type: integer
                        settings:
                          type: object
                          properties:
                            theme:
                              type: string
                    status:
                      oneOf:
                        - type: string
                        - type: integer
                        - type: boolean
                    metadata:
                      type: object
                      properties:
                        tags:
                          type: array
                          items:
                            type: string
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                  code:
                    type: string
    post:
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - email
                - age
              properties:
                name:
                  type: string
                email:
                  type: string
                age:
                  type: integer
                profile:
                  type: object
                  properties:
                    personal:
                      type: object
                      properties:
                        contact:
                          type: object
                          properties:
                            email:
                              type: string
                            phones:
                              type: array
                              items:
                                type: integer
                            address:
                              type: object
                              properties:
                                street:
                                  type: string
                                city:
                                  type: string
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                type: object
                required:
                  - id
                  - name
                  - email
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                  email:
                    type: string
                  createdAt:
                    type: string
                    format: date-time
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
  /users/{id}:
    get:
      operationId: getUserById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                  email:
                    type: string
        '404':
          description: Not Found
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                  code:
                    type: string
    put:
      operationId: updateUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - email
              properties:
                name:
                  type: string
                email:
                  type: string
      responses:
        '200':
          description: Updated
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                  email:
                    type: string
  /users/{id}/profile:
    get:
      operationId: getUserProfile
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  userId:
                    type: integer
                  profile:
                    type: object
                    properties:
                      personal:
                        type: object
                        properties:
                          contact:
                            type: object
                            properties:
                              email:
                                type: string
                              phones:
                                type: array
                                items:
                                  type: integer`

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
	snapshot := formatTestSnapshotWithDiff("complete_spec_with_all_change_types", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
