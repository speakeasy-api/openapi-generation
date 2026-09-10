# typed: true
# frozen_string_literal: true

require_relative "./types"
require_relative "../sdkconfiguration"
require "uri"

module OpenApiSDK
  module SDKHooks
    class TestSDKInitHook < AbstractSDKHook
      extend T::Sig

      class << self
        attr_accessor :init_sdk_version
      end

      sig do
        override
          .params(
            config: SDKConfiguration
          )
          .returns(SDKConfiguration)
      end
      def sdk_init(config:)
        TestSDKInitHook.init_sdk_version = config.sdk_version
        updated_headers = T.must(config.client).headers.dup.merge(
          "Client-Level-Header" => "added by client"
        )
        updated_client = Faraday::Connection.new(
          T.must(config.client).build_exclusive_url,
          headers: updated_headers,
          params: T.must(config.client).params.dup,
          builder: T.must(config.client).builder.dup,
          ssl: T.must(config.client).ssl.dup,
          request: T.must(config.client).options.dup
        )
        config.client = updated_client

        return config
      end
    end

    class TestBeforeRequestHook < AbstractSDKHook
      extend T::Sig

      sig do
        override
          .params(
            hook_ctx: BeforeRequestHookContext,
            request: Faraday::Request
          )
          .returns(Faraday::Request)
      end
      def before_request(hook_ctx:, request:)
        request.headers["Idempotency-Key"] = "some-key"

        case hook_ctx.operation_id
        when "authorizationHeaderModification"
          request.headers["Authorization"] += " modified"
        when "testHooks"
          request.params = request.params.merge("someParam" => "overriddenParam")
        when "testHooksBeforeCreateRequestPaths"
          request.headers["old-pathname"] = URI.parse(request.path).path
        when "hooksCustomUserAgent"
          request.headers["User-Agent"] = "acme-corp/#{hook_ctx.config.sdk_version} acme-corp/#{RUBY_VERSION}"
          request.headers["X-Test-Gen-Version"] = hook_ctx.config.gen_version
          request.headers["X-Test-Doc-Version"] = hook_ctx.config.openapi_doc_version
          request.headers["X-Test-Init-Sdk-Version"] = TestSDKInitHook.init_sdk_version.to_s
        end

        request
      end
    end

    class TestAfterSuccessHook < AbstractSDKHook
      extend T::Sig

      sig do
        override
          .params(
            hook_ctx: AfterSuccessHookContext,
            response: Faraday::Response
          )
          .returns(Faraday::Response)
      end
      def after_success(hook_ctx:, response:)
        case hook_ctx.operation_id
        when "testHooksAfterResponse"
          raise "validation failed"
        end

        response
      end
    end

    class TestAfterErroHook < AbstractSDKHook
      extend T::Sig

      sig do
        override
          .params(
            error: T.nilable(StandardError),
            hook_ctx: AfterErrorHookContext,
            response: T.nilable(Faraday::Response)
          )
          .returns(T.nilable(Faraday::Response))
      end
      def after_error(error:, hook_ctx:, response:)
        case hook_ctx.operation_id
        when "testHooksError"
          raise "expected status code 400" if response && response.status != 400
          raise "special test error case"
        when "statusGetDefaultError"
          if response && response.status == 418
            return Faraday::Response.new(status: 200, response_headers: {}, body: "")
          end
        end

        response
      end
    end
  end
end
