defmodule Hello.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application

  @impl true
  def start(_type, _args) do
    children = [
      HelloWeb.Telemetry,
      {DNSCluster, query: Application.get_env(:hello, :dns_cluster_query) || :ignore},
      {Phoenix.PubSub, name: Hello.PubSub},
      # Start the Finch HTTP client for sending emails
      {Finch, name: Hello.Finch},
      # Start a worker by calling: Hello.Worker.start_link(arg)
      # {Hello.Worker, arg},
      # Start to serve requests, typically the last entry
      HelloWeb.Endpoint
    ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: Hello.Supervisor]
    Supervisor.start_link(children, opts)
  end

  # Tell Phoenix to update the endpoint configuration
  # whenever the application is updated.
  @impl true
  def config_change(changed, _new, removed) do
    HelloWeb.Endpoint.config_change(changed, removed)
    :ok
  end
end
