puts "Hello from Ruby"
puts "Ruby version: #{RUBY_VERSION}"

if RubyVM::YJIT.enabled?
  puts "YJIT: enabled"
else
  puts "YJIT: disabled"
end
