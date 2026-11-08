using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using Finance.Features.BankAccounts;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class BankAccountEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web);

    private readonly HttpClient client = factory.CreateClient();

    public Task InitializeAsync() => factory.ResetDatabaseAsync();

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task CreateReturnsCreatedBankAccount()
    {
        var response = await client.PostAsJsonAsync("/bank-accounts", ValidBankAccount());

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var bankAccount = await response.Content.ReadFromJsonAsync<BankAccountDto>(Json);
        Assert.NotNull(bankAccount);
        Assert.Equal($"/bank-accounts/{bankAccount.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal("Itau operacional", bankAccount.Name);
        Assert.Equal("341", bankAccount.BankCode);
        Assert.Equal("1234", bankAccount.Branch);
        Assert.Equal("56789-0", bankAccount.AccountNumber);
        Assert.Equal(1500.75m, bankAccount.OpeningBalance);
    }

    [Fact]
    public async Task GetReturnsWhatWasCreated()
    {
        var created = await CreateBankAccountAsync();

        var response = await client.GetAsync($"/bank-accounts/{created.Id}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.Equal(created, await response.Content.ReadFromJsonAsync<BankAccountDto>(Json));
    }

    [Fact]
    public async Task GetUnknownReturnsNotFound()
    {
        var response = await client.GetAsync("/bank-accounts/999999");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsDuplicateAccount()
    {
        await CreateBankAccountAsync();

        var response = await client.PostAsJsonAsync("/bank-accounts", ValidBankAccount() with { Name = "Outro nome" });

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<ProblemDetails>(Json);
        Assert.Equal(409, problem?.Status);
    }

    [Fact]
    public async Task CreateAcceptsSameNumberInAnotherBank()
    {
        await CreateBankAccountAsync();

        var response = await client.PostAsJsonAsync("/bank-accounts", ValidBankAccount() with { BankCode = "001" });

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
    }

    [Fact]
    public async Task CreateRejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync("/bank-accounts", new { });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(
            new[] { "Name", "BankCode", "Branch", "AccountNumber", "OpeningBalance" }, problem.Errors.Keys);
    }

    [Theory]
    [InlineData("34")]
    [InlineData("3411")]
    [InlineData("34a")]
    public async Task CreateRejectsInvalidBankCode(string bankCode)
    {
        var response = await client.PostAsJsonAsync("/bank-accounts", ValidBankAccount() with { BankCode = bankCode });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["BankCode"], problem.Errors.Keys);
    }

    private static CreateBankAccountDto ValidBankAccount() =>
        new("Itau operacional", "341", "1234", "56789-0", 1500.75m);

    private async Task<BankAccountDto> CreateBankAccountAsync()
    {
        var response = await client.PostAsJsonAsync("/bank-accounts", ValidBankAccount());
        response.EnsureSuccessStatusCode();
        return (await response.Content.ReadFromJsonAsync<BankAccountDto>(Json))!;
    }
}
