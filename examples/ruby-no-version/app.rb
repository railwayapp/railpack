require 'sinatra'

get '/' do
  "Ruby version: #{RUBY_VERSION}"
end
