# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"
require "active_support"

module OpenApiSDK
  class PaginationLogEntry
    attr_accessor :request_uri, :status_code, :request_body

    def initialize(request_uri, status_code, request_body)
      @request_uri = request_uri
      @status_code = status_code
      @request_body = request_body
    end
  end

  class TestPagination < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_pagination_limit_offset_page_params
      record_test("pagination-limit-offset-page-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      server_limit = 20
      responses = @sdk.pagination.pagination_limit_offset_page_params(page: 1)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(server_limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_union_output_page_params
      record_test("pagination-limit-offset-union-output-page-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      server_limit = 20
      responses = @sdk.pagination.pagination_limit_offset_union_output_page_params(page: 1)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(server_limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_page_body
      record_test("pagination-limit-offset-page-body")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk
        .pagination
        .pagination_limit_offset_page_body(
          request: OpenApiSDK::Models::Shared::LimitOffsetConfig.new(limit: limit, page: 1)
        )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(20 - limit, response.res.result_array.size)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_page_body_nullable
      record_test("pagination-limit-offset-page-body-nullable")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      # first request sends a null body; later pages materialize one with the
      # advanced page (wire sequence enforced by the test service)
      responses = @sdk
        .pagination
        .pagination_limit_offset_page_body_nullable(request: nil)

      pages = []
      responses.each do |response|
        assert(response)
        assert(response.http_meta)
        assert(response.http_meta.response)
        assert_equal(200, response.http_meta.response.status)
        assert(response.res)
        pages << response.res.result_array
      end

      assert_equal(
        [[0, 1, 2, 3, 4, 5, 6], [7, 8, 9, 10, 11, 12, 13], [14, 15, 16, 17, 18, 19]],
        pages
      )
    end

    def test_pagination_limit_offset_deep_outputs_page_body
      record_test("pagination-limit-offset-deep-outputs-page-body")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk
        .pagination
        .pagination_limit_offset_deep_outputs_page_body(
          request: OpenApiSDK::Models::Shared::LimitOffsetConfig.new(limit: limit, page: 1)
        )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(20 - limit, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_offset_params
      record_test("pagination-limit-offset-offset-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk.pagination.pagination_limit_offset_offset_params(limit: limit, offset: 0)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_operator(response.res.result_array.size, :<, limit)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_cursor_params
      record_test("pagination-cursor-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk.pagination.pagination_cursor_params(cursor: -1)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_operator(response.res.result_array.size, :<, limit)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 3
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_cursor_body
      record_test("pagination-cursor-body")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk
        .pagination
        .pagination_cursor_body(
          request: OpenApiSDK::Models::Operations::PaginationCursorBodyRequestBody.new(cursor: -1)
        )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_operator(response.res.result_array.size, :<, limit)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 3
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_cursor_non_numeric
      record_test("pagination-cursor-non-numeric")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      responses = @sdk.pagination.pagination_cursor_non_numeric

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(15, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(5, response.res.result_array.size)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 3
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_cursor_non_numeric_nullable
      record_test("pagination-cursor-non-numeric-nullable")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      responses = @sdk.pagination.pagination_cursor_non_numeric_nullable(cursor: "2")

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(15, response.res.result_array.size)
          assert_equal("17", response.res.cursor)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(2, response.res.result_array.size)
          assert_nil(response.res.cursor)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_url
      record_test("pagination-url")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      responses = @sdk.pagination.pagination_url_params(attempts: 3)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(9, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(6, response.res.result_array.size)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(3, response.res.result_array.size)
        elsif count == 3
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end

      responses = @sdk.pagination.pagination_url_params(attempts: 3, is_reference_path: "true")

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(9, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(6, response.res.result_array.size)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(3, response.res.result_array.size)
        elsif count == 3
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_ambiguous_input
      record_test("pagination-ambiguous-input")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      rb = OpenApiSDK::Models::Operations::PaginationAmbiguousInputRequestBody.new(cursor: -1)
      responses = @sdk.pagination.pagination_ambiguous_input(request_body: rb)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_operator(response.res.result_array.size, :<, limit)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 3
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_optional_security
      record_test("pagination-body-flattened-optional-security")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk.pagination.pagination_body_flattened_optional_security(
        limit: 15,
        offset: 0,
        security: OpenApiSDK::Models::Operations::PaginationBodyFlattenedOptionalSecuritySecurity.new(
          pagination_auth: "test"
        )
      )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(20 - limit, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_body_flattened_with_security
      record_test("pagination-body-flattened-with-security")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk.pagination.pagination_body_flattened_with_security(
        security: OpenApiSDK::Models::Operations::PaginationBodyFlattenedWithSecuritySecurity.new(
          pagination_auth: "test"
        ),
        limit: limit,
        offset: 0
      )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(20 - limit, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_body_wrapped_request
      record_test("pagination-body-wrapped-request")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk.pagination.pagination_body_wrapped_request(
        request: OpenApiSDK::Models::Operations::PaginationBodyWrappedRequestRequest.new(
          limit_offset_config: OpenApiSDK::Models::Shared::LimitOffsetConfig.new(limit: limit, page: 1)
        )
      )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(20 - limit, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_wrapped_optional_body
      record_test("pagination-wrapped-optional-body")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      responses = @sdk.pagination.pagination_wrapped_optional_body(
        request: OpenApiSDK::Models::Operations::PaginationWrappedOptionalBodyRequest.new
      )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(20, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_encapsulated_parameter
      record_test("pagination-encapsulated-parameter")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      rb = OpenApiSDK::Models::Operations::PaginationEncapsulatedParameterRequest.new(cursor: -1)
      responses = @sdk.pagination.pagination_encapsulated_parameter(request: rb)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_operator(response.res.result_array.size, :<, limit)
        elsif count == 2
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(0, response.res.result_array.size)
        elsif count == 3
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_cursor_nullable_limit
      record_test("pagination-cursor-nullable-limit")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      default_limit = 10
      responses = @sdk.pagination.pagination_cursor_nullable_limit

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.pagination_cursor_nullable_limit_next_cursor)
          assert_equal(default_limit, response.pagination_cursor_nullable_limit_next_cursor.results.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.pagination_cursor_nullable_limit_next_cursor)
          assert_equal(default_limit, response.pagination_cursor_nullable_limit_next_cursor.results.size)
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_default_offset_body
      record_test("pagination-limit-offset-default-offset-body")

      # setup faraday client for pagination observability
      connection_options = {
        request: {
          params_encoder: Faraday::FlatParamsEncoder
        }
      }
      client = Faraday.new(**connection_options) do |f|
        f.request(:multipart, {})
        # f.response :logger, nil, { headers: true, bodies: true, errors: true }
        f.request(:instrumentation)
      end

      log = []
      ActiveSupport::Notifications.subscribe("request.faraday") do |_name, _start, _finish, _id, payload|
        log << PaginationLogEntry.new(payload[:url].request_uri, payload[:response].status, payload[:request_body])
      end

      sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, client: client)

      available = 20
      default_offset = 10

      res = sdk.pagination.pagination_limit_offset_default_offset_body(
        request: OpenApiSDK::Models::Shared::LimitOffsetConfigWithDefaults.new(
          # TODO: remove when implementing defaults!
          limit: 15,
          offset: 10
        )
      )

      assert_equal(200, res.http_meta.response.status)
      assert(res.res)
      assert_equal(available - default_offset, res.res.result_array.size)

      assert_equal(1, log.size)
      request_body = log[0].request_body
      refute_nil(request_body)
      assert_includes(request_body, "\"limit\":15")
      assert_includes(request_body, "\"offset\":10")
    end

    def test_pagination_limit_offset_default_offset_params
      record_test("pagination-limit-offset-default-offset-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      offset = 10
      response = @sdk.pagination.pagination_limit_offset_default_offset_params(limit: limit, offset: offset)

      qp = response.http_meta.request.params
      assert_equal(200, response.http_meta.response.status)
      assert_equal(15, qp["limit"].to_i)
      assert_equal(10, qp["offset"].to_i)
    end

    def test_pagination_limit_offset_nil_offset_params
      record_test("pagination-limit-offset-nil-offset-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      available = 20
      default_limit = 20
      responses = @sdk.pagination.pagination_limit_offset_offset_params

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(default_limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(available - default_limit, response.res.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_nil_page_params
      record_test("pagination-limit-offset-nil-page-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      available = 20
      responses = @sdk.pagination.pagination_limit_offset_optional_page_params

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(available, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_empty(response.res.result_array)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_offset_body
      record_test("pagination-limit-offset-offset-body")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      limit = 15
      responses = @sdk
        .pagination
        .pagination_limit_offset_offset_body(
          request: OpenApiSDK::Models::Shared::LimitOffsetConfig.new(limit: limit, offset: 0)
        )

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(limit, response.res.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_operator(response.res.result_array.size, :<, limit)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_zero_page_params
      record_test("pagination-limit-offset-zero-page-params")

      assert_instance_of(OpenApiSDK::SDK, @sdk)

      available = 20
      responses = @sdk.pagination.pagination_limit_offset_optional_page_params(page: 0)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(available, response.res.result_array.size)
          query_params = response.http_meta.request.params
          assert_equal("0", query_params["page"])
        elsif count == 1
          assert(response)
          assert(response.http_meta)
          assert(response.http_meta.response)
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          assert_equal(available, response.res.result_array.size)
          query_params = response.http_meta.request.params
          assert_equal("1", query_params["page"])
        end

        count += 1
      end
    end

    def test_pagination_with_retries
      record_test("pagination-with-retries")
      # setup faraday client for pagination observability
      connection_options = {
        request: {
          params_encoder: Faraday::FlatParamsEncoder
        }
      }
      client = Faraday.new(**connection_options) do |f|
        f.request(:multipart, {})
        # f.response :logger, nil, { headers: true, bodies: true, errors: true }
        f.request(:instrumentation)
      end

      log = []
      ActiveSupport::Notifications.subscribe("request.faraday") do |_name, _start, _finish, _id, payload|
        log << PaginationLogEntry.new(payload[:url].request_uri, payload[:response].status, payload[:request_body])
      end

      available = 20
      total_count = 0
      sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, client: client)

      responses = sdk
        .pagination
        .pagination_with_retries(
          request_id: SecureRandom.uuid,
          fault_settings: "{\"error_code\":503,\"error_count\":3}"
        )

      count = 0
      responses.each do |response|
        if count.zero?
          assert_equal(200, response.http_meta.response.status)
          assert(response.res)
          total_count += response.res.result_array.size
        elsif count == 1
          assert_equal(200, res.http_meta.response.status)
          assert(response.res)
          total_count += res.res.result_array.size
        elsif count == 2
          assert_equal(200, res.http_meta.response.status)
          assert(response.res)
          assert_empty(res.res.result_array)
        elsif count == 3
          assert_nil(res)
        end
      end

      assert_equal(available, total_count)

      assert_equal(200, log[0].status_code)
      refute_includes(log[0].request_uri, "?")
      assert_equal(200, log[1].status_code)
      assert(log[1].request_uri.end_with?("?cursor=14"))
      assert_equal(200, log[2].status_code)
      assert(log[2].request_uri.end_with?("?cursor=19"))
    end
  end
end
