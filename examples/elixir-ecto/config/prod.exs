import Config

# Declare ecto_repos at the application level
config :friends, ecto_repos: [Friends.Repo]

# Runtime production configuration, including reading
# of environment variables, is done in config/runtime.exs.
