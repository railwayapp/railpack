defmodule Friends.MixProject do
  use Mix.Project

  def project do
    [
      app: :friends,
      version: "0.1.0",
      # elixir: "~> 1",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  # Run "mix help compile.app" to learn about applications.
  def application do
    [
      extra_applications: [:logger],
      mod: {Friends.Application, []}
    ]
  end

  # Run "mix help deps" to learn about dependencies.
  defp deps do
    []
  end
end
