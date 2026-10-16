class HomeController < ApplicationController
  def index
    render html: "<h1>Hello from Rails #{Rails.version} on Ruby #{RUBY_VERSION}!</h1>".html_safe, layout: true
  end
end
