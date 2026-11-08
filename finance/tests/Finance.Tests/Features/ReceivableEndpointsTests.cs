using System.Net;
using System.Net.Http.Json;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Finance.Domain;
using Finance.Features.Receivables;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class ReceivableEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web)
    {
        Converters = { new JsonStringEnumConverter() }
    };

    private readonly HttpClient client = factory.CreateClient();

    public Task InitializeAsync() => factory.ResetDatabaseAsync();

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task CreateReturnsCreatedOpenReceivable()
    {
        var response = await client.PostAsJsonAsync("/receivables", ValidReceivable());

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var receivable = await response.Content.ReadFromJsonAsync<ReceivableDto>(Json);
        Assert.NotNull(receivable);
        Assert.Equal($"/receivables/{receivable.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal("Venda de 3 notebooks", receivable.Description);
        Assert.Equal("Acme Ltda", receivable.CustomerName);
        Assert.Equal(4200.50m, receivable.Amount);
        Assert.Equal(new DateOnly(2026, 12, 5), receivable.DueDate);
        Assert.Equal(EntryStatus.Open, receivable.Status);
    }

    [Fact]
    public async Task GetReturnsWhatWasCreated()
    {
        var created = await CreateReceivableAsync();

        var response = await client.GetAsync($"/receivables/{created.Id}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.Equal(created, await response.Content.ReadFromJsonAsync<ReceivableDto>(Json));
    }

    [Fact]
    public async Task GetKeepsCreatedAtOffset()
    {
        var created = await CreateReceivableAsync();

        var body = await client.GetStringAsync($"/receivables/{created.Id}");

        var createdAt = JsonDocument.Parse(body).RootElement.GetProperty("createdAt").GetString();
        Assert.EndsWith("+00:00", createdAt);
    }

    [Fact]
    public async Task GetUnknownReturnsNotFound()
    {
        var response = await client.GetAsync("/receivables/999999");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync("/receivables", new { });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(new[] { "Amount", "Description", "DueDate", "CustomerName" }, problem.Errors.Keys);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-10)]
    public async Task CreateRejectsNonPositiveAmount(decimal amount)
    {
        var response = await client.PostAsJsonAsync("/receivables", ValidReceivable() with { Amount = amount });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Amount"], problem.Errors.Keys);
    }

    [Fact]
    public async Task CreateRejectsMalformedJson()
    {
        var content = new StringContent("""{ "amount": "abc" }""", Encoding.UTF8, "application/json");

        var response = await client.PostAsync("/receivables", content);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    private static CreateReceivableDto ValidReceivable() =>
        new("Venda de 3 notebooks", "Acme Ltda", 4200.50m, new DateOnly(2026, 12, 5));

    private async Task<ReceivableDto> CreateReceivableAsync()
    {
        var response = await client.PostAsJsonAsync("/receivables", ValidReceivable());
        response.EnsureSuccessStatusCode();
        return (await response.Content.ReadFromJsonAsync<ReceivableDto>(Json))!;
    }
}
