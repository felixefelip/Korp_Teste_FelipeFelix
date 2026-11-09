using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using Finance.Features.Customers;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class CustomerEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web);

    private readonly HttpClient client = factory.CreateClient();

    public Task InitializeAsync() => factory.ResetDatabaseAsync();

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task CreateReturnsCreatedCustomer()
    {
        var response = await client.PostAsJsonAsync("/customers", ValidCustomer());

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var customer = await response.Content.ReadFromJsonAsync<CustomerDto>(Json);
        Assert.NotNull(customer);
        Assert.Equal($"/customers/{customer.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal("Acme Ltda", customer.Name);
        Assert.Equal("12345678000195", customer.Document);
    }

    [Fact]
    public async Task GetReturnsWhatWasCreated()
    {
        var created = await (await client.PostAsJsonAsync("/customers", ValidCustomer()))
            .Content.ReadFromJsonAsync<CustomerDto>(Json);

        var fetched = await client.GetFromJsonAsync<CustomerDto>($"/customers/{created!.Id}", Json);

        Assert.Equal(created, fetched);
    }

    [Fact]
    public async Task GetUnknownReturnsNotFound()
    {
        var response = await client.GetAsync("/customers/999999");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsDuplicateDocument()
    {
        (await client.PostAsJsonAsync("/customers", ValidCustomer())).EnsureSuccessStatusCode();

        var response = await client.PostAsJsonAsync("/customers", ValidCustomer() with { Name = "Outro nome" });

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync("/customers", new { });

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
        var response = await client.PostAsJsonAsync("/customers", ValidCustomer() with { Document = document });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Document"], problem.Errors.Keys);
    }

    private static CreateCustomerDto ValidCustomer() => new("Acme Ltda", "12345678000195");
}
