using System.Globalization;
using Microsoft.AspNetCore.Mvc;

namespace csharp.Controllers;

[ApiController]
[Route("[controller]")]
public class WeatherForecastController : ControllerBase
{
    private readonly ILogger<WeatherForecastController> _logger;
    private readonly IConfiguration _configuration;

    public WeatherForecastController(ILogger<WeatherForecastController> logger, IConfiguration configuration)
    {
        _logger = logger;
        _configuration = configuration;
    }

    [HttpGet(Name = "GetWeatherForecast")]
    public IEnumerable<WeatherForecast> Get()
    {
        // Verifies libicu-dev is available for non-invariant globalization in the runtime container
        var culture = CultureInfo.GetCultureInfo("fr-FR");
        // Verifies ASPNETCORE_CONTENTROOT points to /app/out where appsettings.json is published
        var defaultSummary = _configuration["Forecast:DefaultSummary"] ?? "MissingConfig";

        return Enumerable.Range(1, 5).Select(index => new WeatherForecast
        {
            Date = DateTime.Now.AddDays(index),
            TemperatureC = Random.Shared.Next(-20, 55),
            Summary = $"{defaultSummary} ({culture.Name}: {12.5.ToString("F1", culture)})"
        })
        .ToArray();
    }
}
