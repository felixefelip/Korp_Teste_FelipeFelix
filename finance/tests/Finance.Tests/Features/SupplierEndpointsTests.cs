using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using Finance.Features.Suppliers;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class SupplierEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web);

    private readonly HttpClient client = factory.CreateClient();

    public Task InitializeAsync() => factory.ResetDatabaseAsync();

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task CreateReturnsCreatedSupplier()
    {
        var response = await client.PostAsJsonAsync("/suppliers", ValidSupplier());

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var supplier = await response.Content.ReadFromJsonAsync<SupplierDto>(Json);
        Assert.NotNull(supplier);
        Assert.Equal($"/suppliers/{supplier.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal("Imobiliaria Central", supplier.Name);
        Assert.Equal("12345678000195", supplier.Document);
    }

    [Fact]
    public async Task GetReturnsWhatWasCreated()
    {
        var created = await (await client.PostAsJsonAsync("/suppliers", ValidSupplier()))
            .Content.ReadFromJsonAsync<SupplierDto>(Json);

        var fetched = await client.GetFromJsonAsync<SupplierDto>($"/suppliers/{created!.Id}", Json);

        Assert.Equal(created, fetched);
    }

    [Fact]
    public async Task GetUnknownReturnsNotFound()
    {
        var response = await client.GetAsync("/suppliers/999999");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsDuplicateDocument()
    {
        (await client.PostAsJsonAsync("/suppliers", ValidSupplier())).EnsureSuccessStatusCode();

        var response = await client.PostAsJsonAsync("/suppliers", ValidSupplier() with { Name = "Outro nome" });

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync("/suppliers", new { });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(new[] { "Name", "Document" }, problem.Errors.Keys);
    }

    [Theory]
    [InlineData("1234567890")]
    [InlineData("123.456.789-01")]
    public async Task CreateRejectsInvalidDocument(string document)
    {
        var response = await client.PostAsJsonAsync("/suppliers", ValidSupplier() with { Document = document });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Document"], problem.Errors.Keys);
    }

    private static CreateSupplierDto ValidSupplier() => new("Imobiliaria Central", "12345678000195");
}
