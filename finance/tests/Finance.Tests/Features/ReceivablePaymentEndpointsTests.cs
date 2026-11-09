using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using Finance.Domain;
using Finance.Features.Receivables;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class ReceivablePaymentEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web)
    {
        Converters = { new JsonStringEnumConverter() }
    };

    private static readonly DateOnly PaidOn = new(2026, 12, 1);

    private readonly HttpClient client = factory.CreateClient();

    private int bankAccountId;
    private int receivableId;

    public async Task InitializeAsync()
    {
        await factory.ResetDatabaseAsync();
        bankAccountId = await factory.CreateBankAccountAsync();
        receivableId = await CreateReceivableAsync();
    }

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task PartialPaymentReturnsCreatedAndUpdatesReceivable()
    {
        var response = await PayAsync(400m);

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var payment = await response.Content.ReadFromJsonAsync<ReceivablePaymentDto>(Json);
        Assert.NotNull(payment);
        Assert.Equal($"/receivables/{receivableId}/payments/{payment.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal(receivableId, payment.ReceivableId);
        Assert.Equal(400m, payment.Amount);
        Assert.Equal(PaidOn, payment.PaidOn);

        var receivable = await GetReceivableAsync();
        Assert.Equal(400m, receivable.PaidAmount);
        Assert.Equal(600m, receivable.RemainingAmount);
        Assert.Equal(EntryStatus.Open, receivable.Status);
    }

    [Fact]
    public async Task PayingTheRemainingMarksItPaid()
    {
        (await PayAsync(400m)).EnsureSuccessStatusCode();
        (await PayAsync(600m)).EnsureSuccessStatusCode();

        var receivable = await GetReceivableAsync();
        Assert.Equal(1000m, receivable.PaidAmount);
        Assert.Equal(0m, receivable.RemainingAmount);
        Assert.Equal(EntryStatus.Paid, receivable.Status);
    }

    [Fact]
    public async Task RejectsAmountAboveRemaining()
    {
        var response = await PayAsync(1000.01m);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Amount"], problem.Errors.Keys);
        Assert.Equal(0m, (await GetReceivableAsync()).PaidAmount);
    }

    [Fact]
    public async Task RejectsPaymentOnPaidReceivable()
    {
        (await PayAsync(1000m)).EnsureSuccessStatusCode();

        var response = await PayAsync(1m);

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
    }

    [Fact]
    public async Task RejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync($"/receivables/{receivableId}/payments", new { });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(new[] { "Amount", "PaidOn" }, problem.Errors.Keys);
    }

    [Fact]
    public async Task UnknownReceivableReturnsNotFound()
    {
        var payResponse = await client.PostAsJsonAsync(
            "/receivables/999999/payments", new RegisterReceivablePaymentDto(100m, PaidOn));
        var listResponse = await client.GetAsync("/receivables/999999/payments");

        Assert.Equal(HttpStatusCode.NotFound, payResponse.StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, listResponse.StatusCode);
    }

    [Fact]
    public async Task ListReturnsPaymentsOrderedByPaidOn()
    {
        (await PayAsync(100m, new DateOnly(2026, 12, 5))).EnsureSuccessStatusCode();
        (await PayAsync(200m, new DateOnly(2026, 12, 1))).EnsureSuccessStatusCode();

        var payments = await client.GetFromJsonAsync<ReceivablePaymentDto[]>(
            $"/receivables/{receivableId}/payments", Json);

        Assert.NotNull(payments);
        Assert.Equal([200m, 100m], payments.Select(p => p.Amount));
    }

    [Fact]
    public async Task GetPaymentFollowsLocation()
    {
        var response = await PayAsync(400m);
        var created = await response.Content.ReadFromJsonAsync<ReceivablePaymentDto>(Json);

        var fetched = await client.GetFromJsonAsync<ReceivablePaymentDto>(response.Headers.Location, Json);

        Assert.Equal(created, fetched);
    }

    [Fact]
    public async Task GetPaymentOfAnotherReceivableReturnsNotFound()
    {
        var payment = await (await PayAsync(400m)).Content.ReadFromJsonAsync<ReceivablePaymentDto>(Json);
        var otherReceivableId = await CreateReceivableAsync();

        var response = await client.GetAsync($"/receivables/{otherReceivableId}/payments/{payment!.Id}");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    private Task<HttpResponseMessage> PayAsync(decimal amount, DateOnly? paidOn = null) =>
        client.PostAsJsonAsync(
            $"/receivables/{receivableId}/payments", new RegisterReceivablePaymentDto(amount, paidOn ?? PaidOn));

    private async Task<ReceivableDto> GetReceivableAsync() =>
        (await client.GetFromJsonAsync<ReceivableDto>($"/receivables/{receivableId}", Json))!;

    private async Task<int> CreateReceivableAsync()
    {
        var response = await client.PostAsJsonAsync("/receivables",
            new CreateReceivableDto(bankAccountId, "Venda de 3 notebooks", "Acme Ltda", 1000m, new DateOnly(2026, 12, 10)));
        response.EnsureSuccessStatusCode();
        return (await response.Content.ReadFromJsonAsync<ReceivableDto>(Json))!.Id;
    }
}
