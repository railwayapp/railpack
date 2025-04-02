import Config

# Only require DATABASE_URL in certain environments
if config_env() == :prod and not String.starts_with?(System.get_env("MIX_ENV", ""), "assets.") do
  database_url =
    System.get_env("DATABASE_URL")
  config :friends, Friends.Repo,
    url: database_url,
    pool_size: String.to_integer(System.get_env("POOL_SIZE") || "10")
else
  # Development or test config
  config :friends, Friends.Repo,
    url: "ecto://dummy:dummy@localhost/dummy_db"
end

config :friends, ecto_repos: [Friends.Repo]
