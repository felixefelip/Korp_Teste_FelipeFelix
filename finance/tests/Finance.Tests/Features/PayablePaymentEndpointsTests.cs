using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using Finance.Domain;
using Finance.Features.Payables;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class PayablePaymentEndpointsTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web)
    {
        Converters = { new JsonStringEnumConverter() }
    };

    private static readonly DateOnly PaidOn = new(2026, 12, 1);

    private readonly HttpClient client = factory.CreateClient();

    private int bankAccountId;
    private int supplierId;
    private int payableId;

    public async Task InitializeAsync()
    {
        await factory.ResetDatabaseAsync();
        bankAccountId = await factory.CreateBankAccountAsync();
        supplierId = await factory.CreateSupplierAsync();
        payableId = await CreatePayableAsync();
    }

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task PartialPaymentReturnsCreatedAndUpdatesPayable()
    {
        var response = await PayAsync(400m);

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var payment = await response.Content.ReadFromJsonAsync<PayablePaymentDto>(Json);
        Assert.NotNull(payment);
        Assert.Equal($"/payables/{payableId}/payments/{payment.Id}", response.Headers.Location?.OriginalString);
        Assert.Equal(payableId, payment.PayableId);
        Assert.Equal(400m, payment.Amount);
        Assert.Equal(PaidOn, payment.PaidOn);

        var payable = await GetPayableAsync();
        Assert.Equal(400m, payable.PaidAmount);
        Assert.Equal(600m, payable.RemainingAmount);
        Assert.Equal(EntryStatus.Open, payable.Status);
    }

    [Fact]
    public async Task PayingTheRemainingMarksItPaid()
    {
        (await PayAsync(400m)).EnsureSuccessStatusCode();
        (await PayAsync(600m)).EnsureSuccessStatusCode();

        var payable = await GetPayableAsync();
        Assert.Equal(1000m, payable.PaidAmount);
        Assert.Equal(0m, payable.RemainingAmount);
        Assert.Equal(EntryStatus.Paid, payable.Status);
    }

    [Fact]
    public async Task RejectsAmountAboveRemaining()
    {
        var response = await PayAsync(1000.01m);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equal(["Amount"], problem.Errors.Keys);
        Assert.Equal(0m, (await GetPayableAsync()).PaidAmount);
    }

    [Fact]
    public async Task RejectsPaymentOnPaidPayable()
    {
        (await PayAsync(1000m)).EnsureSuccessStatusCode();

        var response = await PayAsync(1m);

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
    }

    [Fact]
    public async Task RejectsMissingFields()
    {
        var response = await client.PostAsJsonAsync($"/payables/{payableId}/payments", new { });

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        Assert.Equivalent(new[] { "Amount", "PaidOn" }, problem.Errors.Keys);
    }

    [Fact]
    public async Task UnknownPayableReturnsNotFound()
    {
        var payResponse = await client.PostAsJsonAsync(
            "/payables/999999/payments", new RegisterPayablePaymentDto(100m, PaidOn));
        var listResponse = await client.GetAsync("/payables/999999/payments");

        Assert.Equal(HttpStatusCode.NotFound, payResponse.StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, listResponse.StatusCode);
    }

    [Fact]
    public async Task ListReturnsPaymentsOrderedByPaidOn()
    {
        (await PayAsync(100m, new DateOnly(2026, 12, 5))).EnsureSuccessStatusCode();
        (await PayAsync(200m, new DateOnly(2026, 12, 1))).EnsureSuccessStatusCode();

        var payments = await client.GetFromJsonAsync<PayablePaymentDto[]>(
            $"/payables/{payableId}/payments", Json);

        Assert.NotNull(payments);
        Assert.Equal([200m, 100m], payments.Select(p => p.Amount));
    }

    [Fact]
    public async Task GetPaymentFollowsLocation()
    {
        var response = await PayAsync(400m);
        var created = await response.Content.ReadFromJsonAsync<PayablePaymentDto>(Json);

        var fetched = await client.GetFromJsonAsync<PayablePaymentDto>(response.Headers.Location, Json);

        Assert.Equal(created, fetched);
    }

    [Fact]
    public async Task GetPaymentOfAnotherPayableReturnsNotFound()
    {
        var payment = await (await PayAsync(400m)).Content.ReadFromJsonAsync<PayablePaymentDto>(Json);
        var otherPayableId = await CreatePayableAsync();

        var response = await client.GetAsync($"/payables/{otherPayableId}/payments/{payment!.Id}");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    private Task<HttpResponseMessage> PayAsync(decimal amount, DateOnly? paidOn = null) =>
        client.PostAsJsonAsync(
            $"/payables/{payableId}/payments", new RegisterPayablePaymentDto(amount, paidOn ?? PaidOn));

    private async Task<PayableDto> GetPayableAsync() =>
        (await client.GetFromJsonAsync<PayableDto>($"/payables/{payableId}", Json))!;

    private async Task<int> CreatePayableAsync()
    {
        var response = await client.PostAsJsonAsync("/payables",
            new CreatePayableDto(bankAccountId, "Aluguel de dezembro", supplierId, 1000m, new DateOnly(2026, 12, 10)));
        response.EnsureSuccessStatusCode();
        return (await response.Content.ReadFromJsonAsync<PayableDto>(Json))!.Id;
    }
}
