defmodule Friends.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application

  @impl true
  def start(_type, _args) do
    IO.puts("Elixir version: #{System.version()}")
    IO.puts("Erlang/OTP version: #{:erlang.system_info(:otp_release)}")

    children = [
      # Starts a worker by calling: Friends.Worker.start_link(arg)
      # {Friends.Worker, arg}
      Friends.Repo,
      {Plug.Cowboy, port: 4000, scheme: :http, plug: Friends.HTTPHandler}
    ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: Friends.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
