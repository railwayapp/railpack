#!/usr/bin/env ruby

require 'bundler/setup'
require 'jemalloc'

if Jemalloc.available?
  puts "JEMALLOC_AVAILABLE"

  # Try to get jemalloc stats to verify it's working
  begin
    stats = Jemalloc.stats
    puts "JEMALLOC_STATS_ACCESSIBLE"
  rescue => e
    puts "JEMALLOC_ERROR: #{e.message}"
    exit 1
  end
else
  puts "JEMALLOC_NOT_AVAILABLE"
  exit 1
end
