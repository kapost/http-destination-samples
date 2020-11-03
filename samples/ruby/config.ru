# frozen_string_literal: true

require "bundler/setup"
Bundler.require(:default)

require "openssl"

class HTTPEndpoint < Sinatra::Base
  SUPPORTED_ACTIONS = ["auth", "publish", "republish"].freeze

  helpers do
    def platform
      @platform ||= request.env["HTTP_X_KAPOST_PLATFORM"].to_s
    end

    def action
      @action ||= request.env["HTTP_X_KAPOST_ACTION"].to_s
    end

    def signature
      @signature ||= request.env["HTTP_X_KAPOST_SIGNATURE"].to_s
    end

    def api_key
      @api_key ||= request.env["HTTP_AUTHORIZATION"].to_s.sub(/^Bearer\s+/i, "")
    end

    def raise_error(code, message)
      halt(code, { "error" => message }.to_json)
    end

    def verify_action!
      raise_error(405, "Action '#{action}' is not supported!") unless SUPPORTED_ACTIONS.include?(action)
    end

    def verify_signature!
      new_signature = [
        "sha256",
        OpenSSL::HMAC.hexdigest(
          OpenSSL::Digest.new("sha256"),
          signature_secret,
          [platform, action, raw_body].join
        )
      ].join("=")

      raise_error(403, "Invalid signature") unless secure_compare(signature, new_signature)
    end

    def verify_api_key!
      raise_error(401, "Invalid API Key") unless secure_compare(api_key, expected_api_key)
    end

    def secure_compare?(a, b)
      Rack::Utils.secure_compare(a, b)
    end

    def expected_api_key
      @expected_api_key ||= ENV.fetch("API_KEY")
    end

    def signature_secret
      @signature_secret ||= ENV.fetch("SIGNATURE_SECRET")
    end

    def raw_body
      @raw_body ||= request.body.read
    end

    def payload
      @payload ||= JSON.parse(raw_body)
    end

    def dump_payload!
      return if production?
      puts JSON.pretty_generate(request.env)
      puts JSON.pretty_generate(payload)
    end

    def production?
      @production ||= (ENV.fetch("RACK_ENV", "dev") == "production")
    end
  end

  before do
    content_type :json
    verify_signature!
    verify_api_key!
    verify_action!
    dump_payload!
  end

  post '/' do
    response = case action
               when "auth"
                 authentication
               when "publish"
                 publish
               when "republish"
                 republish
               else
                 {}
               end

    response.to_json
  end

  private

  def authentication
    # NOTE: the api_key could also be validated against some external service on
    # authentication in the app center and an "error" could be returned if it
    # was incorrect, etc.
    {
      capabilities: {
        any_file: true,
        html: true
      }
    }
  end

  def publish
    {
      metadata: {
        external_id: "abc33",
        published_url: "https://localhost/ruby-test"
      }
    }
  end

  def republish
    external_id = payload.dig("action", "external_id")
    unless secure_compare?(external_id, "abc33")
      raise_error(404, "Cannot republish because external id could not be found")
    end

    {
      metadata: {
        external_id: "abc33",
        published_url: "https://localhost/ruby-test-republish"
      }
    }
  end
end

run HTTPEndpoint
