package namespaces

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

const subSDKCoreConstantCollisionSpec = `openapi: 3.1.0
info:
  title: Widgets
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-retries:
  strategy: backoff
  backoff:
    initialInterval: 500
    maxInterval: 60000
    maxElapsedTime: 3600000
    exponent: 1.5
  statusCodes:
    - 5XX
  retryConnectionErrors: true
tags:
  - name: File
  - name: Time
  - name: Hash
  - name: Class
  - name: Kernel
  - name: Utils
  - name: Date
  - name: String
  - name: Array
  - name: SDKConfiguration
  - name: Widgets
  - name: UtilsSDK
  - name: SDKHooks
paths:
  /files:
    post:
      tags: [File]
      operationId: uploadFile
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required: [file]
              properties:
                file:
                  type: object
                  required: [fileName, content]
                  properties:
                    fileName:
                      type: string
                    content:
                      type: string
                      format: binary
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /time:
    get:
      tags: [Time]
      operationId: getTime
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /hash:
    get:
      tags: [Hash]
      operationId: getHash
      parameters:
        - name: filter
          in: query
          style: deepObject
          schema:
            type: object
            additionalProperties:
              type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /class:
    get:
      tags: [Class]
      operationId: getClass
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /kernel:
    get:
      tags: [Kernel]
      operationId: getKernel
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /utils:
    get:
      tags: [Utils]
      operationId: getUtils
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /date:
    get:
      tags: [Date]
      operationId: getDate
      parameters:
        - name: on
          in: query
          schema:
            type: string
            format: date
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /string:
    get:
      tags: [String]
      operationId: getString
      responses:
        "200":
          description: OK
          content:
            text/plain:
              schema:
                type: string
  /array:
    get:
      tags: [Array]
      operationId: listArray
      parameters:
        - name: page
          in: query
          schema:
            type: integer
      x-speakeasy-pagination:
        type: offsetLimit
        inputs:
          - name: page
            in: parameters
            type: page
        outputs:
          results: $.items
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                    items:
                      $ref: "#/components/schemas/Widget"
  /sdk-configuration:
    get:
      tags: [SDKConfiguration]
      operationId: getSDKConfiguration
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /widgets:
    get:
      tags: [Widgets]
      operationId: listWidgets
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /utils-sdk:
    get:
      tags: [UtilsSDK]
      operationId: getUtilsSDK
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /sdk-hooks:
    get:
      tags: [SDKHooks]
      operationId: getSDKHooks
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
components:
  schemas:
    Widget:
      type: object
      properties:
        id:
          type: string
        createdAt:
          type: string
          format: date-time
          default: "2024-01-01T00:00:00Z"
`

func TestSnapRubySubSDKCoreConstantCollision(t *testing.T) {
	t.Parallel()

	genYaml := `ruby:
  packageName: widgets
`

	expectedSnapshotFiles := []string{
		"lib/widgets.rb",
		"lib/open_api_sdk/sdk.rb",
		"lib/open_api_sdk/utilssdk.rb",
		"lib/open_api_sdk/utilssdksdk.rb",
		"lib/open_api_sdk/sdkhookssdk.rb",
		"lib/open_api_sdk/sdkconfigurationsdk.rb",
		"lib/open_api_sdk/array.rb",
		"lib/open_api_sdk/utils/utils.rb",
		"lib/open_api_sdk/models/components/widget.rb",
	}

	expectedSnapshot := `--- lib/open_api_sdk/array.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

require "faraday"
require "faraday/multipart"
require "faraday/retry"
require "sorbet-runtime"
require "janeway"

require_relative "sdk_hooks/hooks"
require_relative "utils/retries"

module OpenApiSDK
  extend T::Sig

  class Array
    extend T::Sig

    sig { params(sdk_config: SDKConfiguration).void }
    def initialize(sdk_config)
      @sdk_configuration = sdk_config
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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
        page: T.nilable(::Integer),
        retries: T.nilable(Utils::RetryConfig),
        timeout_ms: T.nilable(::Integer),
        http_headers: T.nilable(T::Hash[T.any(::String, ::Symbol), ::String])
      )
        .returns(Models::Operations::ListArrayResponse)
    }
    def list_array(page: nil, retries: nil, timeout_ms: nil, http_headers: nil)

      request = Models::Operations::ListArrayRequest.new(
        page: page
      )
      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/array"
      headers = {}
      headers = T.cast(headers, T::Hash[::String, ::String])
      query_params = Utils.get_query_params(Models::Operations::ListArrayRequest, request, nil)
      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent
      retries ||= @sdk_configuration.retry_config
      retries ||= Utils::RetryConfig.new(
        backoff: Utils::BackoffStrategy.new(
          exponent: 1.5,
          initial_interval: 500,
          max_elapsed_time: 3_600_000,
          max_interval: 60_000
        ),
        retry_connection_errors: true,
        strategy: "backoff"
      )
      retry_options = retries.to_faraday_retry_options(initial_time: ::Time.now)
      retry_options[:retry_statuses] = [500, 501, 502, 503, 504, 505]

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client.dup
      connection.use(Utils::RetryMiddleware, retry_options)

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "listArray",
        security_source: nil
      )

      error = T.let(nil, T.nilable(::StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).get(url) do |req|
          req.headers.merge!(headers)
          req.options.timeout = timeout unless timeout.nil?
          req.params = query_params
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

      rescue ::StandardError => e
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
          obj = Crystalline.unmarshal_json(::JSON.parse(response_data), Models::Operations::ListArrayResponseBody)
          response = Models::Operations::ListArrayResponse.new(
            status_code: http_response.status,
            content_type: content_type,
            raw_response: http_response,
            object: T.unsafe(obj)
          )
          sdk = self

          response.next_page = proc do
            request_page = T.must(!request.page.nil? ? request.page : 1)
            next_page = request_page + 1
            if !response_data
              next nil
            end

            results = Janeway.enum_for("$.items", ::JSON.parse(response_data)).search
            if results.is_a?(::Array)
              results = results[0]
            end

            if results.count.zero?
              next nil
            end

            sdk.list_array(
              page: next_page,
              http_headers: http_headers
            )
          end

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
          :id,
          Crystalline::Nilable.new(::String),
          {'format_json': {'letter_case': ::OpenApiSDK::Utils.field_name("id")}}
        )

        field(
          :created_at,
          Crystalline::Nilable.new(::DateTime),
          {
            'format_json': {
              'letter_case': ::OpenApiSDK::Utils.field_name("createdAt"),
              'decoder': ::OpenApiSDK::Utils.datetime_from_iso_format(true)
            }
          }
        )

        sig { params(id: T.nilable(::String), created_at: T.nilable(::DateTime)).void }
        def initialize(id: nil, created_at: ::DateTime.iso8601("2024-01-01T00:00:00Z"))
          @id = id
          @created_at = created_at
        end

        sig { params(other: T.untyped).returns(T::Boolean) }
        def ==(other)
          return false unless other.is_a?(self.class)
          return false unless @id == other.id
          return false unless @created_at == other.created_at
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

    attr_accessor(
      :file,
      :time,
      :hash,
      :class_,
      :kernel,
      :utils,
      :date,
      :string,
      :array,
      :sdk_configuration_,
      :widgets,
      :utils_sdk,
      :sdk_hooks
    )

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
        timeout_ms: T.nilable(::Integer),
        server_idx: T.nilable(::Integer),
        server_url: T.nilable(::String),
        url_params: T.nilable(T::Hash[::Symbol, ::String])
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
      init_sdks
    end

    sig { void }
    def init_sdks
      @file = File.new(@sdk_configuration)
      @time = Time.new(@sdk_configuration)
      @hash = Hash.new(@sdk_configuration)
      @class_ = Class.new(@sdk_configuration)
      @kernel = Kernel.new(@sdk_configuration)
      @utils = UtilsSDKSDK.new(@sdk_configuration)
      @date = Date.new(@sdk_configuration)
      @string = String.new(@sdk_configuration)
      @array = Array.new(@sdk_configuration)
      @sdk_configuration_ = SDKConfigurationSDK.new(@sdk_configuration)
      @widgets = Widgets.new(@sdk_configuration)
      @utils_sdk = UtilsSDK.new(@sdk_configuration)
      @sdk_hooks = SDKHooksSDK.new(@sdk_configuration)
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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

  end
end


--- lib/open_api_sdk/sdkconfigurationsdk.rb ---
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

  class SDKConfigurationSDK
    extend T::Sig

    sig { params(sdk_config: SDKConfiguration).void }
    def initialize(sdk_config)
      @sdk_configuration = sdk_config
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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
        retries: T.nilable(Utils::RetryConfig),
        timeout_ms: T.nilable(::Integer),
        http_headers: T.nilable(T::Hash[T.any(::String, ::Symbol), ::String])
      )
        .returns(Models::Operations::GetSDKConfigurationResponse)
    }
    def get_sdk_configuration(retries: nil, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/sdk-configuration"
      headers = {}
      headers = T.cast(headers, T::Hash[::String, ::String])
      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent
      retries ||= @sdk_configuration.retry_config
      retries ||= Utils::RetryConfig.new(
        backoff: Utils::BackoffStrategy.new(
          exponent: 1.5,
          initial_interval: 500,
          max_elapsed_time: 3_600_000,
          max_interval: 60_000
        ),
        retry_connection_errors: true,
        strategy: "backoff"
      )
      retry_options = retries.to_faraday_retry_options(initial_time: ::Time.now)
      retry_options[:retry_statuses] = [500, 501, 502, 503, 504, 505]

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client.dup
      connection.use(Utils::RetryMiddleware, retry_options)

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "getSDKConfiguration",
        security_source: nil
      )

      error = T.let(nil, T.nilable(::StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).get(url) do |req|
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

      rescue ::StandardError => e
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
          obj = Crystalline.unmarshal_json(::JSON.parse(response_data), Models::Components::Widget)
          response = Models::Operations::GetSDKConfigurationResponse.new(
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


--- lib/open_api_sdk/sdkhookssdk.rb ---
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

  class SDKHooksSDK
    extend T::Sig

    sig { params(sdk_config: SDKConfiguration).void }
    def initialize(sdk_config)
      @sdk_configuration = sdk_config
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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
        retries: T.nilable(Utils::RetryConfig),
        timeout_ms: T.nilable(::Integer),
        http_headers: T.nilable(T::Hash[T.any(::String, ::Symbol), ::String])
      )
        .returns(Models::Operations::GetSDKHooksResponse)
    }
    def get_sdk_hooks(retries: nil, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/sdk-hooks"
      headers = {}
      headers = T.cast(headers, T::Hash[::String, ::String])
      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent
      retries ||= @sdk_configuration.retry_config
      retries ||= Utils::RetryConfig.new(
        backoff: Utils::BackoffStrategy.new(
          exponent: 1.5,
          initial_interval: 500,
          max_elapsed_time: 3_600_000,
          max_interval: 60_000
        ),
        retry_connection_errors: true,
        strategy: "backoff"
      )
      retry_options = retries.to_faraday_retry_options(initial_time: ::Time.now)
      retry_options[:retry_statuses] = [500, 501, 502, 503, 504, 505]

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client.dup
      connection.use(Utils::RetryMiddleware, retry_options)

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "getSDKHooks",
        security_source: nil
      )

      error = T.let(nil, T.nilable(::StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).get(url) do |req|
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

      rescue ::StandardError => e
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
          obj = Crystalline.unmarshal_json(::JSON.parse(response_data), Models::Components::Widget)
          response = Models::Operations::GetSDKHooksResponse.new(
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


--- lib/open_api_sdk/utils/utils.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

require "cgi"
require "date"
require "base64"
require "stringio"
require "sorbet-runtime"

module OpenApiSDK
  module Utils
    include ::Kernel
    extend T::Sig

    sig { params(val: ::Object).returns(::String) }
    def self.val_to_string(val)
      if val.class.respond_to?(:enums)
        T.unsafe(val).serialize.to_s
      elsif val.is_a?(::DateTime)
        val.strftime("%Y-%m-%dT%H:%M:%S.%NZ").sub(/(\.\d*[1-9])0+Z\z/, "\\1Z").sub(/\.0+Z\z/, "Z")
      else
        val.to_s

      end
    end

    MIME_TYPES = {
      ".json" => "application/json",
      ".xml" => "application/xml",
      ".pdf" => "application/pdf",
      ".zip" => "application/zip",
      ".gz" => "application/gzip",
      ".csv" => "text/csv",
      ".txt" => "text/plain",
      ".html" => "text/html",
      ".htm" => "text/html",
      ".css" => "text/css",
      ".js" => "application/javascript",
      ".png" => "image/png",
      ".jpg" => "image/jpeg",
      ".jpeg" => "image/jpeg",
      ".gif" => "image/gif",
      ".svg" => "image/svg+xml",
      ".webp" => "image/webp",
      ".mp3" => "audio/mpeg",
      ".mp4" => "video/mp4",
      ".yaml" => "application/yaml",
      ".yml" => "application/yaml"
    }.freeze

    sig { params(filename: ::String).returns(::String) }
    def self.mime_type_for_filename(filename)
      ext = ::File.extname(filename).downcase
      MIME_TYPES.fetch(ext, "application/octet-stream")
    end

    sig do
      params(metadata: T::Hash[::Symbol, ::String], field_name: ::String, obj: ::Object)
        .returns(T::Hash[::Symbol, T::Array[::String]])
    end
    def self._get_serialized_params(metadata, field_name, obj)
      params = {}

      serialization = metadata.fetch(:serialization, "")
      params[metadata.fetch(:field_name, field_name)] = [obj.to_json] if serialization == "json"

      params
    end

    sig { params(status: ::Integer).returns(T::Boolean) }
    def self.error_status?(status)
      status_major = status / 100
      return true if status_major == 4
      return true if status_major == 5

      false
    end

    sig { params(content_type: ::String, pattern: ::String).returns(T::Boolean) }
    def self.match_content_type(content_type, pattern)
      return true if ["*", "*/*"].include?(pattern)

      # Media types are case-insensitive (RFC 9110 section 8.3.1).
      pattern = pattern.strip.downcase
      return true if content_type.strip.downcase == pattern

      pieces = content_type.split(";")
      pieces.each do |piece|
        return true if pattern == piece.strip.downcase
      end

      false
    end

    sig { params(status_code: ::Integer, status_codes: T::Array[::String]).returns(T::Boolean) }
    def self.match_status_code(status_code, status_codes)
      return true if status_codes.include?("default")
      status_code = status_code.to_s
      status_codes.each do |code|
        return true if code == status_code
        return true if code.downcase.end_with?("xx") && status_code[0] == code[0]
      end

      false
    end

    sig { params(optional: T::Boolean).returns(T.proc.params(s: ::String).returns(T.nilable(::DateTime))) }
    def self.datetime_from_iso_format(optional)
      ::Kernel.lambda do |s|
        return nil if optional && s.nil?

        return ::DateTime.strptime(s, "%Y-%m-%dT%H:%M:%S.%NZ")
      end
    end

    sig { params(optional: T::Boolean).returns(T.proc.params(s: ::String).returns(T.nilable(::Date))) }
    def self.date_from_iso_format(optional)
      ::Kernel.lambda do |s|
        return nil if optional && s.nil?

        return ::Date.iso8601(s)
      end
    end

    sig do
      params(enum_type: T.class_of(T::Enum), optional: T::Boolean)
        .returns(T.nilable(T.proc.params(s: ::String).returns(T.nilable(T::Enum))))
    end
    def self.enum_from_string(enum_type, optional)
      ::Kernel.lambda do |s|
        return nil if optional && s.nil?

        return enum_type.deserialize(s)
      end
    end

    sig do
      params(enum_type: ::Module, optional: T::Boolean)
        .returns(T.proc.params(s: ::String).returns(T.untyped))
    end
    def self.open_enum_from_string(enum_type, optional)
      ::Kernel.lambda do |s|
        return nil if optional && s.nil?

        return T.unsafe(enum_type).deserialize(s)
      end
    end

    sig { params(name: ::String).returns(T.proc.returns(::String)) }
    def self.field_name(name)
      T.let(proc { name }, T.proc.returns(::String))
    end

    sig { params(field: ::Crystalline::MetadataFields::Field, data_class: ::Object).returns(::Object) }
    def self.parse_field(field, data_class)
      field_metadata = field.metadata[:metadata_string]
      return nil if field_metadata.nil?

      field_value = data_class.send(field.name)
      return nil if field_value.nil?

      field_value
    end

    sig do
      params(form: T::Array[T::Array[T.any(::String, T::Array[T.nilable(::String)])]])
        .returns(T::Hash[::Symbol, ::Object])
    end
    def self.encode_form(form)
      payload = {}
      form.each do |field_name, field|
        field = T.must(field)
        if field.length == 2
          if field[0].nil?
            # Handle multiple values for the same field name (arrays)
            if payload.key?(field_name)
              # Convert to array if not already
              unless payload[field_name].is_a?(::Array)
                payload[field_name] = [payload[field_name]]
              end

              payload[field_name] << field[1]
            else
              payload[field_name] = field[1]
            end
          else
            # Handle file uploads
            file_part = Faraday::Multipart::FilePart.new(
              ::StringIO.new(T.must(field[1])),
              mime_type_for_filename(T.must(field[0])),
              field[0]
            )

            # Handle multiple files for the same field name (arrays)
            if payload.key?(field_name)
              unless payload[field_name].is_a?(::Array)
                payload[field_name] = [payload[field_name]]
              end

              payload[field_name] << file_part
            else
              payload[field_name] = file_part
            end
          end
        elsif field.length == 3
          param_part = Faraday::Multipart::ParamPart.new(field[1], field[2])

          # Handle multiple values for the same field name (arrays)
          if payload.key?(field_name)
            unless payload[field_name].is_a?(::Array)
              payload[field_name] = [payload[field_name]]
            end

            payload[field_name] << param_part
          else
            payload[field_name] = param_part
          end
        end
      end

      payload
    end

    sig {
      params(
        param_name: ::Symbol,
        value: ::Object,
        param_type: ::String,
        gbls: T.nilable(T::Hash[::Symbol, T::Hash[::Symbol, T::Hash[::Symbol, ::Object]]])
      )
        .returns(::Object)
    }
    def self._populate_from_globals(param_name, value, param_type, gbls)
      if value.nil? && !gbls.nil?
        global_value = gbls.dig(:parameters, param_type.to_sym, param_name.to_sym)
        value = global_value if !global_value.nil?
      end

      value
    end

  end
end


--- lib/open_api_sdk/utilssdk.rb ---
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

  class UtilsSDK
    extend T::Sig

    sig { params(sdk_config: SDKConfiguration).void }
    def initialize(sdk_config)
      @sdk_configuration = sdk_config
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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
        retries: T.nilable(Utils::RetryConfig),
        timeout_ms: T.nilable(::Integer),
        http_headers: T.nilable(T::Hash[T.any(::String, ::Symbol), ::String])
      )
        .returns(Models::Operations::GetUtilsSDKResponse)
    }
    def get_utils_sdk(retries: nil, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/utils-sdk"
      headers = {}
      headers = T.cast(headers, T::Hash[::String, ::String])
      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent
      retries ||= @sdk_configuration.retry_config
      retries ||= Utils::RetryConfig.new(
        backoff: Utils::BackoffStrategy.new(
          exponent: 1.5,
          initial_interval: 500,
          max_elapsed_time: 3_600_000,
          max_interval: 60_000
        ),
        retry_connection_errors: true,
        strategy: "backoff"
      )
      retry_options = retries.to_faraday_retry_options(initial_time: ::Time.now)
      retry_options[:retry_statuses] = [500, 501, 502, 503, 504, 505]

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client.dup
      connection.use(Utils::RetryMiddleware, retry_options)

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "getUtilsSDK",
        security_source: nil
      )

      error = T.let(nil, T.nilable(::StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).get(url) do |req|
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

      rescue ::StandardError => e
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
          obj = Crystalline.unmarshal_json(::JSON.parse(response_data), Models::Components::Widget)
          response = Models::Operations::GetUtilsSDKResponse.new(
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


--- lib/open_api_sdk/utilssdksdk.rb ---
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

  class UtilsSDKSDK
    extend T::Sig

    sig { params(sdk_config: SDKConfiguration).void }
    def initialize(sdk_config)
      @sdk_configuration = sdk_config
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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
        retries: T.nilable(Utils::RetryConfig),
        timeout_ms: T.nilable(::Integer),
        http_headers: T.nilable(T::Hash[T.any(::String, ::Symbol), ::String])
      )
        .returns(Models::Operations::GetUtilsResponse)
    }
    def get_utils(retries: nil, timeout_ms: nil, http_headers: nil)

      url, params = @sdk_configuration.get_server_details
      base_url = Utils.template_url(url, params)
      url = "#{base_url}/utils"
      headers = {}
      headers = T.cast(headers, T::Hash[::String, ::String])
      headers["Accept"] = "application/json"
      headers["user-agent"] = @sdk_configuration.user_agent
      retries ||= @sdk_configuration.retry_config
      retries ||= Utils::RetryConfig.new(
        backoff: Utils::BackoffStrategy.new(
          exponent: 1.5,
          initial_interval: 500,
          max_elapsed_time: 3_600_000,
          max_interval: 60_000
        ),
        retry_connection_errors: true,
        strategy: "backoff"
      )
      retry_options = retries.to_faraday_retry_options(initial_time: ::Time.now)
      retry_options[:retry_statuses] = [500, 501, 502, 503, 504, 505]

      timeout = (timeout_ms.to_f / 1000) unless timeout_ms.nil?
      timeout ||= @sdk_configuration.timeout

      connection = @sdk_configuration.client.dup
      connection.use(Utils::RetryMiddleware, retry_options)

      hook_ctx = SDKHooks::HookContext.new(
        config: @sdk_configuration,
        base_url: base_url,
        oauth2_scopes: nil,
        operation_id: "getUtils",
        security_source: nil
      )

      error = T.let(nil, T.nilable(::StandardError))
      http_response = T.let(nil, T.nilable(Faraday::Response))

      begin
        http_response = T.must(connection).get(url) do |req|
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

      rescue ::StandardError => e
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
          obj = Crystalline.unmarshal_json(::JSON.parse(response_data), Models::Components::Widget)
          response = Models::Operations::GetUtilsResponse.new(
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


--- lib/widgets.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  autoload :SDK, "open_api_sdk/sdk"
  autoload :File, "open_api_sdk/file"
  autoload :Time, "open_api_sdk/time"
  autoload :Hash, "open_api_sdk/hash"
  autoload :Class, "open_api_sdk/class_"
  autoload :Kernel, "open_api_sdk/kernel"
  autoload :UtilsSDKSDK, "open_api_sdk/utilssdksdk"
  autoload :Date, "open_api_sdk/date"
  autoload :String, "open_api_sdk/string"
  autoload :Array, "open_api_sdk/array"
  autoload :SDKConfigurationSDK, "open_api_sdk/sdkconfigurationsdk"
  autoload :Widgets, "open_api_sdk/widgets"
  autoload :UtilsSDK, "open_api_sdk/utilssdk"
  autoload :SDKHooksSDK, "open_api_sdk/sdkhookssdk"

  module Models
    autoload :Components, "open_api_sdk/models/components"
    autoload :Operations, "open_api_sdk/models/operations"
    autoload :Errors, "open_api_sdk/models/errors"
    autoload :Callbacks, "open_api_sdk/models/callbacks"
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


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         subSDKCoreConstantCollisionSpec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}

const subSDKReservedNameNestingSpec = `openapi: 3.1.0
info:
  title: Widgets
  version: 1.0.0
servers:
  - url: https://api.example.com
tags:
  - name: Utils
paths:
  /utils:
    get:
      tags: [Utils]
      operationId: getUtils
      responses:
        "200":
          description: OK
  /admin/utils:
    get:
      x-speakeasy-group: admin.utils
      operationId: getAdminUtils
      responses:
        "200":
          description: OK
  /admin/sdk-configuration:
    get:
      x-speakeasy-group: admin.SDKConfiguration
      operationId: getAdminSDKConfiguration
      responses:
        "200":
          description: OK
  /admin/sdk-hooks:
    get:
      x-speakeasy-group: admin.SDKHooks
      operationId: getAdminSDKHooks
      responses:
        "200":
          description: OK
`

func TestSnapRubySubSDKReservedNameNesting(t *testing.T) {
	t.Parallel()

	genYaml := `generation:
  sdkClassName: UtilsSDK
ruby:
  packageName: widgets
`

	expectedSnapshotFiles := []string{
		"lib/widgets.rb",
		"lib/open_api_sdk/utilssdk.rb",
		"lib/open_api_sdk/admin.rb",
	}

	expectedSnapshot := `--- lib/open_api_sdk/admin.rb ---
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

  class Admin
    extend T::Sig
    attr_accessor :utils, :sdk_configuration_, :sdk_hooks

    sig { params(sdk_config: SDKConfiguration).void }
    def initialize(sdk_config)
      @sdk_configuration = sdk_config
      init_sdks
    end

    sig { void }
    def init_sdks
      @utils = ::OpenApiSDK::AdminUtils.new(@sdk_configuration)
      @sdk_configuration_ = ::OpenApiSDK::SDKConfigurationSDK.new(@sdk_configuration)
      @sdk_hooks = ::OpenApiSDK::SDKHooksSDK.new(@sdk_configuration)
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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
  end
end


--- lib/open_api_sdk/utilssdk.rb ---
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

  class UtilsSDK
    extend T::Sig

    attr_accessor :utils, :admin

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
        timeout_ms: T.nilable(::Integer),
        server_idx: T.nilable(::Integer),
        server_url: T.nilable(::String),
        url_params: T.nilable(T::Hash[::Symbol, ::String])
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
      init_sdks
    end

    sig { void }
    def init_sdks
      @utils = UtilsSDKSDK.new(@sdk_configuration)
      @admin = Admin.new(@sdk_configuration)
    end

    sig {
      params(base_url: ::String, url_variables: T.nilable(T::Hash[::Symbol, T.any(::String, T::Enum)])).returns(
        ::String
      )
    }
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

  end
end


--- lib/widgets.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# typed: true
# frozen_string_literal: true

module OpenApiSDK
  autoload :UtilsSDK, "open_api_sdk/utilssdk"
  autoload :UtilsSDKSDK, "open_api_sdk/utilssdksdk"
  autoload :Admin, "open_api_sdk/admin"
  autoload :AdminUtils, "open_api_sdk/admin_utils"
  autoload :SDKConfigurationSDK, "open_api_sdk/sdkconfigurationsdk"
  autoload :SDKHooksSDK, "open_api_sdk/sdkhookssdk"

  module Models
    autoload :Components, "open_api_sdk/models/components"
    autoload :Operations, "open_api_sdk/models/operations"
    autoload :Errors, "open_api_sdk/models/errors"
    autoload :Callbacks, "open_api_sdk/models/callbacks"
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


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         subSDKReservedNameNestingSpec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}

const coreConstantExampleValuesSpec = `openapi: 3.1.0
info:
  title: Widgets
  version: 1.0.0
servers:
  - url: https://api.example.com
tags:
  - name: File
paths:
  /files:
    post:
      tags: [File]
      operationId: uploadFile
      parameters:
        - name: on
          in: query
          required: true
          schema:
            type: string
            format: date
            example: "2024-01-02"
        - name: since
          in: query
          required: true
          schema:
            type: string
            format: date-time
            example: "2024-01-02T03:04:05Z"
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required: [file]
              properties:
                file:
                  type: string
                  format: binary
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
`

func TestSnapRubyCoreConstantExampleValues(t *testing.T) {
	t.Parallel()

	genYaml := `generation:
  tests:
    generateTests: true
    generateNewTests: true
ruby:
  packageName: widgets
`

	expectedSnapshotFiles := []string{
		"USAGE.md",
		"test/file_test.rb",
	}

	expectedSnapshot := `--- USAGE.md ---
<!-- Start SDK Example Usage [usage] -->
` + "`" + `` + "`" + `` + "`" + `ruby
require "widgets"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.file.upload_file(
  on: Date.parse("2024-01-02"),
  since: DateTime.iso8601("2024-01-02T03:04:05Z"),
  body: Models::Operations::UploadFileRequestBody.new(
    file: Models::Operations::File.new(
      file_name: "example.file",
      content: File.binread("example.file")
    )
  )
)

unless res.object.nil?
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `
<!-- End SDK Example Usage [usage] -->

--- test/file_test.rb ---
# Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

# frozen_string_literal: true

require_relative "../lib/widgets"
require_relative "test_client"
require_relative "common_helper_test"
require "minitest/autorun"
require "minitest/focus"

module OpenApiSDK
  class TestFile < Minitest::Test

    def test_file_upload_file
      test_http_client = create_test_http_client("uploadFile")

      s = ::OpenApiSDK::SDK.new(
        server_url: ::ENV.fetch("TEST_SERVER_URL", "http://localhost:18080"),
        client: test_http_client
      )
      refute_nil(s)

      res = s.file.upload_file(
        on: ::Date.parse("2024-01-02"),
        since: ::DateTime.iso8601("2024-01-02T03:04:05Z"),
        body: Models::Operations::UploadFileRequestBody.new(
          file: Models::Operations::File.new(
            file_name: "example.file",
            content: ::File.binread(".speakeasy/testfiles/example.file")
          )
        )
      )
      refute_nil(res)
      assert_equal(200, res.status_code)
      refute_nil(res.object)
      assert_equal(Models::Operations::UploadFileResponseBody.new, res.object)
    end

  end
end


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         coreConstantExampleValuesSpec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
