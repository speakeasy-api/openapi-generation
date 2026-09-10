package securityusage

const specOrSecurity = `openapi: 3.1.0
info:
  title: Security Hoisting with Simple security schemes (OR)
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - auth1: []
  - auth2: []
paths:
  /op0:
    get:
      operationId: globalSecurity
      responses:
        '204':
          description: No Content
  /op1:
    get:
      operationId: auth1Hoisted
      security:
        - auth1: []
      responses:
        '204':
          description: No Content
  /op2:
    get:
      operationId: auth2Hoisted
      security:
        - auth2: []
      responses:
        '204':
          description: No Content
  /op3:
    get:
      operationId: auth2Preferred
      security:
        - auth2: []
        - auth1: []
      responses:
        '204':
          description: No Content
  /op4:
    get:
      operationId: opLevelClientCredentials
      security:
        - oauth2cc:
            - read:data
      responses:
        '204':
          description: No Content
components:
  securitySchemes:
    auth1:
      type: apiKey
      in: header
      name: X-Auth-1
    auth2:
      type: http
      scheme: basic
    oauth2cc:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://auth.example.com/oauth/token
          scopes:
            read:data: Read data
            write:data: Write data`

const specAndSecurity = `openapi: 3.1.0
info:
  title: Security Hoisting with Composite Security Requirement (AND)
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - auth1: []
    auth2: []
paths:
  /op0:
    get:
      operationId: globalSecurity
      responses:
        '204':
          description: No Content
  /op1:
    get:
      operationId: andAuthHoisted
      security:
        - auth1: []
          auth2: []
      responses:
        '204':
          description: No Content
  /op2:
    get:
      operationId: opLevelAndAuth
      security:
        - auth2: []
          auth3: []
      responses:
        '204':
          description: No Content
components:
  securitySchemes:
    auth1:
      type: apiKey
      in: header
      name: X-Auth-1
    auth2:
      type: http
      scheme: basic
    auth3:
      type: http
      scheme: bearer`

const specMixedSecurity = `openapi: 3.1.0
info:
  title: Security Hoisting with Mixed Security Requirements (AND/OR)
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - authA1: []
    authA2: []
  - authB: []
  - authC: [read]
paths:
  /op0:
    get:
      operationId: globalSecurity
      responses:
        '204':
          description: No Content
  /opA:
    get:
      operationId: option1Hoisted
      security:
        - authA1: []
          authA2: []
      responses:
        '204':
          description: No Content
  /opB:
    get:
      operationId: option2Hoisted
      security:
        - authB: []
      responses:
        '204':
          description: No Content
  /opBC:
    get:
      operationId: option1NotAllowed
      security:
        - authC: [write]
        - authB: []
      responses:
        '204':
          description: No Content
  /not/hoisted:
    get:
      operationId: opLevelMixedAuth
      security:
        - authA1: []
          authB: []
        - authC: [read]
      responses:
        '204':
          description: No Content
components:
  securitySchemes:
    authA1:
      type: apiKey
      in: header
      name: X-Auth-1
    authA2:
      type: http
      scheme: basic
    authB:
      type: http
      scheme: bearer
      x-speakeasy-example: <YOUR_JWT>
    authC:
      type: oauth2
      flows:
        implicit:
          authorizationUrl: /oauth2/authorize
          scopes:
            read: Read access
            write: Write access`

const specGlobalOptional = `openapi: 3.1.0
info:
  title: Optional Global Security
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - authA: []
  - authB: []
  - {}
paths:
  /op0:
    get:
      operationId: globalSecurityOptional
      responses:
        '204':
          description: No Content
  /op1:
    get:
      operationId: hoistedOptional
      security:
        - authB: []
        - {}
      responses:
        '204':
          description: No Content
  /op2:
    get:
      operationId: opLevelRequired
      security:
        - authB: []
      responses:
        '204':
          description: No Content
components:
  securitySchemes:
    authA:
      type: apiKey
      in: header
      name: X-Auth-1
    authB:
      type: http
      scheme: bearer
      x-speakeasy-example: <YOUR_JWT>`

const specOpOptional = `openapi: 3.1.0
info:
  title: Optional, Hoisted Operation Security
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - auth1: []
paths:
  /op0:
    get:
      operationId: globalSecurityRequired
      responses:
        '204':
          description: No Content
  /op1:
    get:
      operationId: opLevelOptional
      security:
        - auth1: []
        - {}
      responses:
        '204':
          description: No Content
  /op2:
    get:
      operationId: disableAuthWithEmptyObject
      security:
        - {}
      responses:
        '204':
          description: No Content
  /op3:
    get:
      operationId: disableAuthWithEmptyArray
      security: []
      responses:
        '204':
          description: No Content
components:
  securitySchemes:
    auth1:
      type: apiKey
      in: header
      name: X-Auth-1`

const specNoGlobalWithOp = `openapi: 3.1.0
info:
  title: No Global Auth API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /op0:
    get:
      operationId: globalSecurityHoisted
      responses:
        '204':
          description: No Content
  /op1:
    get:
      operationId: mostCommonSchemeHoisted
      security:
        - authA: []
      responses:
        '204':
          description: No Content
  /op2:
    get:
      operationId: opLevelSecurity
      security:
        - authA: []
        - authB: []
      responses:
        '204':
          description: No Content
components:
  securitySchemes:
    authA:
      type: apiKey
      in: header
      name: X-Auth-1
    authB:
      type: http
      scheme: bearer`
