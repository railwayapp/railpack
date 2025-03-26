class WelcomeController < ApplicationController
  def index
    message = "Hello from Rails"
    render plain: message, content_type: "text/plain"
  end
end
