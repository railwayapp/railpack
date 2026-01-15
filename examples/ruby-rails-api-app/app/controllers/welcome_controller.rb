class WelcomeController < ApplicationController
  def index
    message = "Hello from Ruby: #{RUBY_VERSION}"
    render plain: message, content_type: "text/plain"
  end
end
