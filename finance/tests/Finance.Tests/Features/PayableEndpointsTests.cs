using System.Net;
using System.Net.Http.Json;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Finance.Domain;
using Finance.Features.Payables;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class PayableEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web)
    {
        Converters = { new JsonStringEnumConverter() }
    };

    private readonly HttpClient client = factory.CreateClient();

    private int bankAccountId;
    private int supplierId;

    public async Task InitializeAsync()
    {
        await factory.ResetDatabaseAsync();
        bankAccountId = await factory.CreateBankAccountAsync();
        supplierId = await factory.CreateSupplierAsync();
    }

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task CreateReturnsCreatedOpenPayable()
    {
        var response = await client.PostAsJsonAsync("/payables", ValidPayable());

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var payable = await response.Content.ReadFromJsonAsync<PayableDto>(Json);
        Assert.NotNull(payable);
        Assert.Equal($"/payables/{payable.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal("Aluguel de dezembro", payable.Description);
        Assert.Equal(supplierId, payable.SupplierId);
        Assert.Equal(bankAccountId, payable.BankAccountId);
        Assert.Equal(4200.50m, payable.Amount);
        Assert.Equal(0m, payable.PaidAmount);
        Assert.Equal(4200.50m, payable.RemainingAmount);
        Assert.Equal(new DateOnly(2026, 12, 5), payable.DueDate);
        Assert.Equal(EntryStatus.Open, payable.Status);
    }

    [Fact]
    public async Task GetReturnsWhatWasCreated()
    {
        var created = await CreatePayableAsync();

        var response = await client.GetAsync($"/payables/{created.Id}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.Equal(created, await response.Content.ReadFromJsonAsync<PayableDto>(Json));
    }

    [Fact]
    public async Task GetKeepsCreatedAtOffset()
    {
        var created = await CreatePayableAsync();

        var body = await client.GetStringAsync($"/payables/{created.Id}");

        var createdAt = JsonDocument.Parse(body).RootElement.GetProperty("createdAt").GetString();
        Assert.EndsWith("+00:00", createdAt);
    }

    [Fact]
    public async Task GetUnknownReturnsNotFound()
    {
        var response = await client.GetAsync("/payables/999999");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync("/payables", new { });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(new[] { "Amount", "BankAccountId", "Description", "DueDate", "SupplierId" }, problem.Errors.Keys);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-10)]
    public async Task CreateRejectsNonPositiveAmount(decimal amount)
    {
        var response = await client.PostAsJsonAsync("/payables", ValidPayable() with { Amount = amount });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Amount"], problem.Errors.Keys);
    }

    [Fact]
    public async Task CreateRejectsUnknownBankAccount()
    {
        var response = await client.PostAsJsonAsync("/payables", ValidPayable() with { BankAccountId = bankAccountId + 1000 });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["BankAccountId"], problem.Errors.Keys);
    }

    [Fact]
    public async Task CreateRejectsUnknownSupplier()
    {
        var response = await client.PostAsJsonAsync("/payables", ValidPayable() with { SupplierId = supplierId + 1000 });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["SupplierId"], problem.Errors.Keys);
    }

    [Fact]
    public async Task CreateReportsEveryUnknownReference()
    {
        var response = await client.PostAsJsonAsync(
            "/payables", ValidPayable() with { BankAccountId = bankAccountId + 1000, SupplierId = supplierId + 1000 });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(new[] { "BankAccountId", "SupplierId" }, problem.Errors.Keys);
    }

    [Fact]
    public async Task CreateRejectsMalformedJson()
    {
        var content = new StringContent("""{ "amount": "abc" }""", Encoding.UTF8, "application/json");

        var response = await client.PostAsync("/payables", content);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    private CreatePayableDto ValidPayable() =>
        new(bankAccountId, "Aluguel de dezembro", supplierId, 4200.50m, new DateOnly(2026, 12, 5));

    private async Task<PayableDto> CreatePayableAsync()
    {
        var response = await client.PostAsJsonAsync("/payables", ValidPayable());
        response.EnsureSuccessStatusCode();
        return (await response.Content.ReadFromJsonAsync<PayableDto>(Json))!;
    }
}
