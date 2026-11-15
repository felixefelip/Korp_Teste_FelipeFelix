using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using Finance.Domain;
using Finance.Features.Stakeholders;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class StakeholderEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web)
    {
        Converters = { new JsonStringEnumConverter() }
    };

    private readonly HttpClient client = factory.CreateClient();

    public Task InitializeAsync() => factory.ResetDatabaseAsync();

    public Task DisposeAsync() => Task.CompletedTask;

    [Theory]
    [InlineData(StakeholderType.Customer)]
    [InlineData(StakeholderType.Supplier)]
    public async Task CreateReturnsCreatedStakeholder(StakeholderType type)
    {
        var response = await client.PostAsJsonAsync("/stakeholders", ValidStakeholder() with { Type = type }, Json);

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var stakeholder = await response.Content.ReadFromJsonAsync<StakeholderDto>(Json);
        Assert.NotNull(stakeholder);
        Assert.Equal($"/stakeholders/{stakeholder.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal("Acme Ltda", stakeholder.Name);
        Assert.Equal("12345678000195", stakeholder.Document);
        Assert.Equal(type, stakeholder.Type);
    }

    [Fact]
    public async Task CreateReturnsTypeAsString()
    {
        var response = await client.PostAsJsonAsync("/stakeholders", ValidStakeholder(), Json);

        var body = await response.Content.ReadFromJsonAsync<JsonElement>();
        Assert.Equal("Customer", body.GetProperty("type").GetString());
    }

    [Fact]
    public async Task GetReturnsWhatWasCreated()
    {
        var created = await (await client.PostAsJsonAsync("/stakeholders", ValidStakeholder(), Json))
            .Content.ReadFromJsonAsync<StakeholderDto>(Json);

        var fetched = await client.GetFromJsonAsync<StakeholderDto>($"/stakeholders/{created!.Id}", Json);

        Assert.Equal(created, fetched);
    }

    [Fact]
    public async Task GetUnknownReturnsNotFound()
    {
        var response = await client.GetAsync("/stakeholders/999999");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsDuplicateDocumentOfSameType()
    {
        (await client.PostAsJsonAsync("/stakeholders", ValidStakeholder(), Json)).EnsureSuccessStatusCode();

        var response = await client.PostAsJsonAsync("/stakeholders", ValidStakeholder() with { Name = "Outro nome" }, Json);

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
    }

    [Fact]
    public async Task CreateAcceptsSameDocumentWithOtherType()
    {
        (await client.PostAsJsonAsync("/stakeholders", ValidStakeholder(), Json)).EnsureSuccessStatusCode();

        var response = await client.PostAsJsonAsync(
            "/stakeholders", ValidStakeholder() with { Type = StakeholderType.Supplier }, Json);

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync("/stakeholders", new { });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(new[] { "Name", "Document", "Type" }, problem.Errors.Keys);
    }

    [Theory]
    [InlineData("1234567890")]
    [InlineData("123.456.789-01")]
    public async Task CreateRejectsInvalidDocument(string document)
    {
        var response = await client.PostAsJsonAsync("/stakeholders", ValidStakeholder() with { Document = document }, Json);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Document"], problem.Errors.Keys);
    }

    [Fact]
    public async Task CreateRejectsUndefinedNumericType()
    {
        var response = await client.PostAsJsonAsync(
            "/stakeholders", new { Name = "Acme Ltda", Document = "12345678000195", Type = 99 });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Type"], problem.Errors.Keys);
    }

    [Fact]
    public async Task CreateRejectsUnknownTypeName()
    {
        var response = await client.PostAsJsonAsync(
            "/stakeholders", new { Name = "Acme Ltda", Document = "12345678000195", Type = "Partner" });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    private static CreateStakeholderDto ValidStakeholder() =>
        new("Acme Ltda", "12345678000195", StakeholderType.Customer);
}
