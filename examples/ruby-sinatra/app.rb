require 'sinatra'

puts "Hello from Sinatra"
puts "Ruby version: #{RUBY_VERSION}"

get '/' do
  'Choo Choo! Welcome to your Sinatra server 🚅'
end
