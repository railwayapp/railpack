require 'sinatra'

get '/' do
  "Hello from Ruby #{RUBY_VERSION}!"
end
