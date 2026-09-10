package namespaces

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapRubyNamespaceCollisionStress stress-tests Ruby namespace collision detection
// by using x-speakeasy-model-namespace values that match common internal SDK module/directory
// names (utils, models, errors, operations, hooks, etc.). Each "Widget" model is placed into
// a namespace that is likely to collide with an internal Ruby module, and we verify
// the generator produces valid, non-colliding output.
//
// Uses the same spec as TestSnapPyNamespaceCollisionStress plus Ruby-specific namespaces
// (shared, sdk_hooks, security, server_variables).
func TestSnapRubyNamespaceCollisionStress(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Namespace Collision Stress Test
  version: 1.0.0
  description: >
    Stress-tests namespace collision detection by using x-speakeasy-model-namespace
    values that match common internal SDK package/module names (utils, models, errors,
    types, hooks, sdk, etc.). Each Widget model is placed into a namespace that is
    likely to collide with an internal package. The generator must detect these
    collisions and produce valid, non-colliding output.
servers:
  - url: https://api.example.com
paths:
  /ns/utils:
    post:
      operationId: utils
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/utils_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/utils_Widget"
  /ns/lib:
    post:
      operationId: lib
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/lib_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/lib_Widget"
  /ns/sdk:
    post:
      operationId: sdk
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/sdk_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/sdk_Widget"
  /ns/models:
    post:
      operationId: models
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/models_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/models_Widget"
  /ns/operations:
    post:
      operationId: operations
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/operations_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/operations_Widget"
  /ns/errors:
    post:
      operationId: errors
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/errors_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/errors_Widget"
        "4XX":
          description: Error
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string
                  code:
                    type: integer
  /ns/funcs:
    post:
      operationId: funcs
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/funcs_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/funcs_Widget"
  /ns/types:
    post:
      operationId: types
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/types_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/types_Widget"
  /ns/timeout:
    post:
      operationId: timeout
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/timeout_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/timeout_Widget"
  /ns/retry:
    post:
      operationId: retry
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/retry_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/retry_Widget"
  /ns/retries:
    post:
      operationId: retries
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/retries_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/retries_Widget"
  /ns/hooks:
    post:
      operationId: hooks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/hooks_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/hooks_Widget"
  /ns/webhooks:
    post:
      operationId: webhooks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/webhooks_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/webhooks_Widget"
  /ns/callbacks:
    post:
      operationId: callbacks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/callbacks_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/callbacks_Widget"
  /ns/enums:
    post:
      operationId: enums
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/enums_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/enums_Widget"
  /ns/unions:
    post:
      operationId: unions
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/unions_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/unions_Widget"
  /ns/core:
    post:
      operationId: core
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/core_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/core_Widget"
  /ns/test:
    post:
      operationId: test
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Test_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Test_Widget"
  /ns/async:
    post:
      operationId: async
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Async_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Async_Widget"
  /ns/auth:
    post:
      operationId: auth
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Auth_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Auth_Widget"
  /ns/docs:
    post:
      operationId: docs
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Docs_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Docs_Widget"
  /ns/components:
    post:
      operationId: components
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/components_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/components_Widget"
  /ns/shared:
    post:
      operationId: shared
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/shared_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/shared_Widget"
  /ns/sdk_hooks:
    post:
      operationId: sdk_hooks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/sdk_hooks_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/sdk_hooks_Widget"
  /ns/security:
    post:
      operationId: security
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/security_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/security_Widget"
  /ns/server_variables:
    post:
      operationId: server_variables
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/server_variables_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/server_variables_Widget"
components:
  schemas:
    utils_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: utils
      type: object
      properties:
        utils:
          type: string
    lib_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: lib
      type: object
      properties:
        lib:
          type: string
    sdk_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: sdk
      type: object
      properties:
        sdk:
          type: string
    models_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: models
      type: object
      properties:
        models:
          type: string
    operations_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: operations
      type: object
      properties:
        operations:
          type: string
    errors_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: errors
      type: object
      properties:
        errors:
          type: string
    funcs_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: funcs
      type: object
      properties:
        funcs:
          type: string
    types_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: types
      type: object
      properties:
        types:
          type: string
    timeout_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: timeout
      type: object
      properties:
        timeout:
          type: string
    retry_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: retry
      type: object
      properties:
        retry:
          type: string
    retries_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: retries
      type: object
      properties:
        retries:
          type: string
    hooks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: hooks
      type: object
      properties:
        hooks:
          type: string
    webhooks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: webhooks
      type: object
      properties:
        webhooks:
          type: string
    callbacks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: callbacks
      type: object
      properties:
        callbacks:
          type: string
    enums_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: enums
      type: object
      properties:
        enums:
          type: string
    unions_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: unions
      type: object
      properties:
        unions:
          type: string
    core_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: core
      type: object
      properties:
        core:
          type: string
    Test_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Test
      type: object
      properties:
        test:
          type: string
    Async_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Async
      type: object
      properties:
        async:
          type: string
    Auth_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Auth
      type: object
      properties:
        auth:
          type: string
    Docs_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Docs
      type: object
      properties:
        docs:
          type: string
    components_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: components
      type: object
      properties:
        components:
          type: string
    shared_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: shared
      type: object
      properties:
        shared:
          type: string
    sdk_hooks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: sdk_hooks
      type: object
      properties:
        sdk_hooks:
          type: string
    security_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: security
      type: object
      properties:
        security:
          type: string
    server_variables_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: server_variables
      type: object
      properties:
        server_variables:
          type: string`

	genYaml := `ruby:
  packageName: nstest
`

	expectedSnapshotFiles := []string{
		"lib/nstest.rb",
		"lib/open_api_sdk/sdk.rb",
		"lib/open_api_sdk/models/**/*.rb",
	}

	expectedSnapshot := `--- lib/nstest.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  autoload :SDK, "open_api_sdk/sdk"

  module Models
    autoload :Components, "open_api_sdk/models/components"
    autoload :Operations, "open_api_sdk/models/operations"
    autoload :Errors, "open_api_sdk/models/errors"
    autoload :Callbacks, "open_api_sdk/models/callbacks"
    autoload :Webhooks, "open_api_sdk/models/webhooks"
    autoload :Utils, "open_api_sdk/models/utils"
    autoload :Lib, "open_api_sdk/models/lib"
    autoload :SDK, "open_api_sdk/models/sdk"
    autoload :Models, "open_api_sdk/models/models"
    autoload :Funcs, "open_api_sdk/models/funcs"
    autoload :Types, "open_api_sdk/models/types"
    autoload :Timeout, "open_api_sdk/models/timeout"
    autoload :Retry, "open_api_sdk/models/retry"
    autoload :Retries, "open_api_sdk/models/retries"
    autoload :Hooks, "open_api_sdk/models/hooks"
    autoload :Enums, "open_api_sdk/models/enums"
    autoload :Unions, "open_api_sdk/models/unions"
    autoload :Core, "open_api_sdk/models/core"
    autoload :Test, "open_api_sdk/models/test"
    autoload :Async, "open_api_sdk/models/async"
    autoload :Auth, "open_api_sdk/models/auth"
    autoload :Docs, "open_api_sdk/models/docs"
    autoload :Shared, "open_api_sdk/models/shared"
    autoload :SDKHooks, "open_api_sdk/models/sdk_hooks"
    autoload :Security, "open_api_sdk/models/security"
    autoload :ServerVariables, "open_api_sdk/models/server_variables"
  end
end

require_relative "open_api_sdk/utils/utils"
require_relative "open_api_sdk/utils/request_bodies"
require_relative "open_api_sdk/utils/query_params"
require_relative "open_api_sdk/utils/forms"
require_relative "open_api_sdk/utils/headers"
require_relative "open_api_sdk/utils/url"
require_relative "open_api_sdk/utils/security"
require_relative "crystalline"
require_relative "open_api_sdk/sdkconfiguration"


--- lib/open_api_sdk/models/async/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Async

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :async,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("async")}}
        )

        sig { params(async: T.nilable(::String)).void }
        def initialize(async: nil)
          @async = async
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @async == other.async
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/auth/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Auth

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :auth,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("auth")}}
        )

        sig { params(auth: T.nilable(::String)).void }
        def initialize(auth: nil)
          @auth = auth
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @auth == other.auth
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/callbacks/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Callbacks

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :callbacks,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("callbacks")}}
        )

        sig { params(callbacks: T.nilable(::String)).void }
        def initialize(callbacks: nil)
          @callbacks = callbacks
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @callbacks == other.callbacks
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/components/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Components

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :components,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("components")}}
        )

        sig { params(components: T.nilable(::String)).void }
        def initialize(components: nil)
          @components = components
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @components == other.components
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/core/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Core

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :core,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("core")}}
        )

        sig { params(core: T.nilable(::String)).void }
        def initialize(core: nil)
          @core = core
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @core == other.core
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/docs/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Docs

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :docs,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("docs")}}
        )

        sig { params(docs: T.nilable(::String)).void }
        def initialize(docs: nil)
          @docs = docs
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @docs == other.docs
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/enums/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Enums

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :enums,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("enums")}}
        )

        sig { params(enums: T.nilable(::String)).void }
        def initialize(enums: nil)
          @enums = enums
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @enums == other.enums
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/errors/apierror.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Errors

      class APIError < StandardError
        include Crystalline::MetadataFields
        extend T::Sig

        field :body, T.nilable(::String), {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("body")}}
        field :raw_response, T.nilable(Faraday::Response), {}
        field(
          :status_code,
          T.nilable(::Integer),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("status_code")}}
        )

        sig { params(status_code: ::Integer, body: ::String, raw_response: Faraday::Response).void }
        def initialize(status_code:, body:, raw_response:)
          @status_code = status_code
          @body = body
          @raw_response = raw_response
        end

        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @status_code == other.status_code
          return false unless @body == other.body
          return false unless @raw_response == other.raw_response
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/errors/clienterror.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Errors
      # Error
      class ClientError < StandardError
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :message,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("message")}}
        )

        field(
          :code,
          Crystalline::Nilable.new(::Integer),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("code")}}
        )
        # Raw HTTP response; suitable for custom response parsing
        field(
          :raw_response,
          Crystalline::Nilable.new(::Faraday::Response),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("-")}}
        )

        sig {
          params(message: T.nilable(::String), code: T.nilable(::Integer), raw_response: T.nilable(::Faraday::Response))
            .void
        }
        def initialize(message: nil, code: nil, raw_response: nil)
          @message = message
          @code = code
          @raw_response = raw_response
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @message == other.message
          return false unless @code == other.code
          return false unless @raw_response == other.raw_response
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/errors/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Errors

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :errors,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("errors")}}
        )

        sig { params(errors: T.nilable(::String)).void }
        def initialize(errors: nil)
          @errors = errors
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @errors == other.errors
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/funcs/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Funcs

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :funcs,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("funcs")}}
        )

        sig { params(funcs: T.nilable(::String)).void }
        def initialize(funcs: nil)
          @funcs = funcs
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @funcs == other.funcs
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/hooks/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Hooks

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :hooks,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("hooks")}}
        )

        sig { params(hooks: T.nilable(::String)).void }
        def initialize(hooks: nil)
          @hooks = hooks
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @hooks == other.hooks
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/lib/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Lib

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :lib,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("lib")}}
        )

        sig { params(lib: T.nilable(::String)).void }
        def initialize(lib: nil)
          @lib = lib
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @lib == other.lib
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/models/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Models

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :models,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("models")}}
        )

        sig { params(models: T.nilable(::String)).void }
        def initialize(models: nil)
          @models = models
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @models == other.models
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/async_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class AsyncResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Async::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Async::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/auth_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class AuthResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Auth::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Auth::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/callbacks_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class CallbacksResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Callbacks::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Callbacks::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/components_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class ComponentsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Components::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Components::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/core_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class CoreResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Core::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Core::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/docs_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class DocsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Docs::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Docs::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/enums_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class EnumsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Enums::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Enums::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/errors_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class ErrorsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Errors::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Errors::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/funcs_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class FuncsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Funcs::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Funcs::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/hooks_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class HooksResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Hooks::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Hooks::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/lib_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class LibResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Lib::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Lib::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/models_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class ModelsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Models::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Models::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/operations_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class OperationsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Operations::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Operations::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/retries_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class RetriesResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Retries::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Retries::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/retry_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class RetryResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Retry::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Retry::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/sdk_hooks_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class SDKHooksResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::SDKHooks::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::SDKHooks::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/sdk_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class SDKResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::SDK::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::SDK::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/security_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class SecurityResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Security::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Security::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/server_variables_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class ServerVariablesResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::ServerVariables::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::ServerVariables::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/shared_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class SharedResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Shared::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Shared::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/test_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class TestResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Test::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Test::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/timeout_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class TimeoutResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Timeout::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Timeout::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/types_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class TypesResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Types::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Types::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/unions_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class UnionsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Unions::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Unions::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/utils_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class UtilsResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Utils::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Utils::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/webhooks_response.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class WebhooksResponse
        extend T::Sig
        include Crystalline::MetadataFields

        # HTTP response content type for this operation
        field :content_type, ::String
        # HTTP response status code for this operation
        field :status_code, ::Integer
        # Raw HTTP response; suitable for custom response parsing
        field :raw_response, ::Faraday::Response
        # OK
        field :widget, Crystalline::Nilable.new(::OpenApiSDK::Models::Webhooks::Widget)

        sig {
          params(
            content_type: ::String,
            status_code: ::Integer,
            raw_response: ::Faraday::Response,
            widget: T.nilable(::OpenApiSDK::Models::Webhooks::Widget)
          )
            .void
        }
        def initialize(content_type:, status_code:, raw_response:, widget: nil)
          @content_type = content_type
          @status_code = status_code
          @raw_response = raw_response
          @widget = widget
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @content_type == other.content_type
          return false unless @status_code == other.status_code
          return false unless @raw_response == other.raw_response
          return false unless @widget == other.widget
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/operations/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Operations

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :operations,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("operations")}}
        )

        sig { params(operations: T.nilable(::String)).void }
        def initialize(operations: nil)
          @operations = operations
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @operations == other.operations
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/retries/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Retries

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :retries,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("retries")}}
        )

        sig { params(retries: T.nilable(::String)).void }
        def initialize(retries: nil)
          @retries = retries
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @retries == other.retries
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/retry/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Retry

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :retry_,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("retry")}}
        )

        sig { params(retry_: T.nilable(::String)).void }
        def initialize(retry_: nil)
          @retry_ = retry_
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @retry_ == other.retry_
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/sdk/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module SDK

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :sdk,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("sdk")}}
        )

        sig { params(sdk: T.nilable(::String)).void }
        def initialize(sdk: nil)
          @sdk = sdk
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @sdk == other.sdk
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/sdk_hooks/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module SDKHooks

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :sdk_hooks,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("sdk_hooks")}}
        )

        sig { params(sdk_hooks: T.nilable(::String)).void }
        def initialize(sdk_hooks: nil)
          @sdk_hooks = sdk_hooks
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @sdk_hooks == other.sdk_hooks
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/security/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Security

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :security,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("security")}}
        )

        sig { params(security: T.nilable(::String)).void }
        def initialize(security: nil)
          @security = security
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @security == other.security
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/server_variables/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module ServerVariables

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :server_variables,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("server_variables")}}
        )

        sig { params(server_variables: T.nilable(::String)).void }
        def initialize(server_variables: nil)
          @server_variables = server_variables
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @server_variables == other.server_variables
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/shared/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Shared

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :shared,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("shared")}}
        )

        sig { params(shared: T.nilable(::String)).void }
        def initialize(shared: nil)
          @shared = shared
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @shared == other.shared
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/test/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Test

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :test,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("test")}}
        )

        sig { params(test: T.nilable(::String)).void }
        def initialize(test: nil)
          @test = test
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @test == other.test
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/timeout/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Timeout

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :timeout,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("timeout")}}
        )

        sig { params(timeout: T.nilable(::String)).void }
        def initialize(timeout: nil)
          @timeout = timeout
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @timeout == other.timeout
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/types/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Types

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :types,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("types")}}
        )

        sig { params(types: T.nilable(::String)).void }
        def initialize(types: nil)
          @types = types
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @types == other.types
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/unions/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Unions

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :unions,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("unions")}}
        )

        sig { params(unions: T.nilable(::String)).void }
        def initialize(unions: nil)
          @unions = unions
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @unions == other.unions
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/utils/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Utils

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :utils,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("utils")}}
        )

        sig { params(utils: T.nilable(::String)).void }
        def initialize(utils: nil)
          @utils = utils
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @utils == other.utils
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/models/webhooks/widget.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  module Models
    module Webhooks

      class Widget
        extend T::Sig
        include Crystalline::MetadataFields

        field(
          :webhooks,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("webhooks")}}
        )

        sig { params(webhooks: T.nilable(::String)).void }
        def initialize(webhooks: nil)
          @webhooks = webhooks
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @webhooks == other.webhooks
          true
        end
      end
    end
  end
end


--- lib/open_api_sdk/sdk.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

require "faraday"
require "faraday/multipart"
require "faraday/retry"
require "sorbet-runtime"

require_relative "sdk_hooks/hooks"
require_relative "utils/retries"

module OpenApiSDK
  extend T::Sig

  class SDK
    extend T::Sig

    # Instantiates the SDK, configuring it with the provided parameters.
    #
    # @param client [Faraday::Connection, nil] The faraday HTTP client to use for all operations
    # @param retry_config [::OpenApiSDK::Utils::RetryConfig, nil] The retry configuration to use for all operations
    # @param timeout_ms [Integer, nil] Request timeout in milliseconds for all operations
    # @param server_idx [Integer, nil] The index of the server to use for all operations
    # @param server_url [String, nil] The server URL to use for all operations
    # @param url_params [Hash{Symbol => String}, nil] Parameters to optionally template the server URL with
    sig do
      params(
        client: T.nilable(Faraday::Connection),
        retry_config: T.nilable(::OpenApiSDK::Utils::RetryConfig),
        timeout_ms: T.nilable(Integer),
        server_idx: T.nilable(Integer),
        server_url: T.nilable(String),
        url_params: T.nilable(T::Hash[Symbol, String])
      )
        .void
    end
    def initialize(client: nil, retry_config: nil, timeout_ms: nil, server_idx: nil, server_url: nil, url_params: nil)

      connection_options = {
        request: {
          params_encoder: Faraday::FlatParamsEncoder
        }
      }
      connection_options[:request][:timeout] = (timeout_ms.to_f / 1000) unless timeout_ms.nil?

      client ||= Faraday.new(**connection_options) do |f|
        f.request(:multipart, {flat_encode: true})
        # f.response :logger, nil, { headers: true, bodies: true, errors: true }
      end

      if !server_url.nil?
        if !url_params.nil?
          server_url = Utils.template_url(server_url, url_params)
        end
      end

      server_idx = 0 if server_idx.nil?
      hooks = SDKHooks::Hooks.new
      @sdk_configuration = SDKConfiguration.new(
        client,
        hooks,
        retry_config,
        timeout_ms,
        server_url,
        server_idx
      )
      @sdk_configuration = hooks.sdk_init(config: @sdk_configuration)
    end

    sig { params(base_url: String, url_variables: T.nilable(T::Hash[Symbol, T.any(String, T::Enum)])).returns(String) }
    def get_url(base_url:, url_variables: nil)
      sd_base_url, sd_options = @sdk_configuration.get_server_details

      if base_url.nil?
        base_url = sd_base_url
      end

      if url_variables.nil?
        url_variables = sd_options
      end

      return Utils.template_url(base_url, url_variables)
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Utils::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::UtilsResponse)
    }
    def utils(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/utils"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "utils",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Utils::Widget)
          response = ::OpenApiSDK::Models::Operations::UtilsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Lib::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::LibResponse)
    }
    def lib(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/lib"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "lib",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Lib::Widget)
          response = ::OpenApiSDK::Models::Operations::LibResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::SDK::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::SDKResponse)
    }
    def sdk(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/sdk"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "sdk",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::SDK::Widget)
          response = ::OpenApiSDK::Models::Operations::SDKResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Models::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::ModelsResponse)
    }
    def models(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/models"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "models",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Models::Widget)
          response = ::OpenApiSDK::Models::Operations::ModelsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Operations::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::OperationsResponse)
    }
    def operations(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/operations"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "operations",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Operations::Widget)
          response = ::OpenApiSDK::Models::Operations::OperationsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Errors::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::ErrorsResponse)
    }
    def errors(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/errors"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "errors",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Errors::Widget)
          response = ::OpenApiSDK::Models::Operations::ErrorsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Errors::ClientError)
          obj.raw_response = http_response
          raise obj
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Funcs::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::FuncsResponse)
    }
    def funcs(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/funcs"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "funcs",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Funcs::Widget)
          response = ::OpenApiSDK::Models::Operations::FuncsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Types::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::TypesResponse)
    }
    def types(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/types"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "types",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Types::Widget)
          response = ::OpenApiSDK::Models::Operations::TypesResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Timeout::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::TimeoutResponse)
    }
    def timeout(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/timeout"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "timeout",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Timeout::Widget)
          response = ::OpenApiSDK::Models::Operations::TimeoutResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Retry::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::RetryResponse)
    }
    def retry(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/retry"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "retry",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Retry::Widget)
          response = ::OpenApiSDK::Models::Operations::RetryResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Retries::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::RetriesResponse)
    }
    def retries(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/retries"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "retries",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Retries::Widget)
          response = ::OpenApiSDK::Models::Operations::RetriesResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Hooks::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::HooksResponse)
    }
    def hooks(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/hooks"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "hooks",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Hooks::Widget)
          response = ::OpenApiSDK::Models::Operations::HooksResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Webhooks::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::WebhooksResponse)
    }
    def webhooks(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/webhooks"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "webhooks",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Webhooks::Widget)
          response = ::OpenApiSDK::Models::Operations::WebhooksResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Callbacks::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::CallbacksResponse)
    }
    def callbacks(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/callbacks"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "callbacks",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Callbacks::Widget)
          response = ::OpenApiSDK::Models::Operations::CallbacksResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Enums::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::EnumsResponse)
    }
    def enums(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/enums"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "enums",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Enums::Widget)
          response = ::OpenApiSDK::Models::Operations::EnumsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Unions::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::UnionsResponse)
    }
    def unions(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/unions"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "unions",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Unions::Widget)
          response = ::OpenApiSDK::Models::Operations::UnionsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Core::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::CoreResponse)
    }
    def core(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/core"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "core",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Core::Widget)
          response = ::OpenApiSDK::Models::Operations::CoreResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Test::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::TestResponse)
    }
    def test(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/test"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "test",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Test::Widget)
          response = ::OpenApiSDK::Models::Operations::TestResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Async::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::AsyncResponse)
    }
    def async(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/async"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "async",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Async::Widget)
          response = ::OpenApiSDK::Models::Operations::AsyncResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Auth::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::AuthResponse)
    }
    def auth(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/auth"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "auth",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Auth::Widget)
          response = ::OpenApiSDK::Models::Operations::AuthResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Docs::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::DocsResponse)
    }
    def docs(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/docs"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "docs",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Docs::Widget)
          response = ::OpenApiSDK::Models::Operations::DocsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Components::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::ComponentsResponse)
    }
    def components(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/components"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "components",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Components::Widget)
          response = ::OpenApiSDK::Models::Operations::ComponentsResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Shared::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::SharedResponse)
    }
    def shared(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/shared"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "shared",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Shared::Widget)
          response = ::OpenApiSDK::Models::Operations::SharedResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::SDKHooks::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::SDKHooksResponse)
    }
    def sdk_hooks(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/sdk_hooks"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "sdk_hooks",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::SDKHooks::Widget)
          response = ::OpenApiSDK::Models::Operations::SDKHooksResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::Security::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::SecurityResponse)
    }
    def security(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/security"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "security",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::Security::Widget)
          response = ::OpenApiSDK::Models::Operations::SecurityResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end

    sig {
      params(
        request: ::OpenApiSDK::Models::ServerVariables::Widget,
        timeout_ms: T.nilable(Integer),
        http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])
      )
        .returns(::OpenApiSDK::Models::Operations::ServerVariablesResponse)
    }
    def server_variables(request:, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/ns/server_variables"
      headers = {}
      headers = T.cast(headers, T::Hash[String, String])
      req_content_type, data, form = Utils.serialize_request_body(request, false, false, :request, :json)
      headers["content-type"] = req_content_type
      raise StandardError, "request body is required" if data.nil? && form.nil?

      if form && !form.empty?
        body = Utils.encode_form(form)
      elsif Utils.match_content_type(req_content_type, "application/x-www-form-urlencoded")
        body = URI.encode_www_form(T.cast(data, T::Hash[Symbol, Object]))
      else
        body = data
      end

      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "server_variables",
        security_source: nil
      )

      error = T.let(nil, T.nilable(StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).post(url) do |req|
          req.body = body
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          http_headers&.each do |key, value|
            req.headers[key.to_s] = value
          end

          @sdk_configuration.hooks.before_request(
            hook_ctx: SDKHooks::BeforeRequestHookContext.new(
              hook_ctx: hook_ctx
            ),
            request: req
          )
        end

      rescue StandardError => e
        error = e
      ensure
        if http_response.nil? || Utils.error_status?(http_response.status)
          http_response = @sdk_configuration.hooks.after_error(
            error: error,
            hook_ctx: SDKHooks::AfterErrorHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        else
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
        end

        if http_response.nil?
          raise error if !error.nil?
          raise "no response"
        end
      end

      content_type = http_response.headers.fetch("Content-Type", "application/octet-stream")
      if Utils.match_status_code(http_response.status, ["200"])
        if Utils.match_content_type(content_type, "application/json")
          http_response = @sdk_configuration.hooks.after_success(
            hook_ctx: SDKHooks::AfterSuccessHookContext.new(
              hook_ctx: hook_ctx
            ),
            response: http_response
          )
          response_data = http_response.env.response_body
          obj = Crystalline.unmarshal_json(JSON.parse(response_data), ::OpenApiSDK::Models::ServerVariables::Widget)
          response = ::OpenApiSDK::Models::Operations::ServerVariablesResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            widget: T.unsafe(obj)
          )

          return response
        else
          raise(
            ::OpenApiSDK::Models::Errors::APIError.new(
              status_code: http_response.status,
              body: http_response.env.response_body,
              raw_response: http_response
            ),
            "Unknown content type received"
          )
        end
      elsif Utils.match_status_code(http_response.status, ["4XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      elsif Utils.match_status_code(http_response.status, ["5XX"])
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "API error occurred"
        )
      else
        raise(
          ::OpenApiSDK::Models::Errors::APIError.new(
            status_code: http_response.status,
            body: http_response.env.response_body,
            raw_response: http_response
          ),
          "Unknown status code received"
        )

      end
    end
  end
end


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
