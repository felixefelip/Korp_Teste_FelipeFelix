using System.Net;
using System.Net.Http.Json;
using System.Text;
using System.Text.Json;
using Finance.Data;
using Finance.Features.Suppliers;
using Finance.Tests.Infrastructure;
using Microsoft.AspNetCore.Http;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace Finance.Tests.Features;

[Collection(ApiCollection.Name)]
public class SupplierImportTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private const string Route = "/suppliers/import";

    private static readonly JsonSerializerOptions Json = new(JsonSerializerDefaults.Web);
    private static readonly string[] Headers = ["Name", "Document"];

    private readonly HttpClient client = factory.CreateClient();

    public Task InitializeAsync() => factory.ResetDatabaseAsync();

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task ImportsEveryRow()
    {
        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers,
            ["Acme Ltda", "12345678000195"],
            ["Maria Silva", "12345678901"]));

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        var suppliers = await response.Content.ReadFromJsonAsync<SupplierDto[]>(Json);
        Assert.NotNull(suppliers);
        Assert.Equal(["Acme Ltda", "Maria Silva"], suppliers.Select(c => c.Name));
        Assert.Equal(["12345678000195", "12345678901"], suppliers.Select(c => c.Document));
        Assert.Equal(suppliers[0], await client.GetFromJsonAsync<SupplierDto>($"/suppliers/{suppliers[0].Id}", Json));
    }

    [Fact]
    public async Task NormalizesMaskedAndNumericDocuments()
    {
        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers,
            ["Acme Ltda", "12.345.678/0001-95"],
            ["Maria Silva", "123.456.789-01"],
            ["Joao Souza", 1234567890d],
            ["Beta SA", 1234567000188d]));

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        var suppliers = await response.Content.ReadFromJsonAsync<SupplierDto[]>(Json);
        Assert.NotNull(suppliers);
        Assert.Equal(
            ["12345678000195", "12345678901", "01234567890", "01234567000188"],
            suppliers.Select(c => c.Document));
    }

    [Fact]
    public async Task AcceptsColumnsInAnyOrderAndCase()
    {
        var response = await client.PostAsync(Route, Spreadsheet.Upload(["document", "NAME"],
            ["12345678000195", "Acme Ltda"]));

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        var suppliers = await response.Content.ReadFromJsonAsync<SupplierDto[]>(Json);
        Assert.Equal("Acme Ltda", Assert.Single(suppliers!).Name);
    }

    [Fact]
    public async Task RejectsWholeFileWhenARowIsInvalid()
    {
        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers,
            ["Acme Ltda", "12345678000195"],
            ["Maria Silva", "123"],
            ["", "12345678901"]));

        var errors = await ErrorsAsync(response);
        Assert.Equivalent(new[] { "Row 3", "Row 4" }, errors.Keys);
        Assert.Equal(0, await CountAsync());
    }

    [Fact]
    public async Task RejectsDocumentRepeatedInFile()
    {
        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers,
            ["Acme Ltda", "12345678000195"],
            ["Acme Filial", "12.345.678/0001-95"]));

        var errors = await ErrorsAsync(response);
        Assert.Contains("row 2", Assert.Single(errors["Row 3"]));
    }

    [Fact]
    public async Task RejectsDocumentAlreadyRegistered()
    {
        await factory.CreateSupplierAsync("12345678000195");

        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers,
            ["Acme Ltda", "12345678000195"]));

        var errors = await ErrorsAsync(response);
        Assert.Equal(["Row 2"], errors.Keys);
        Assert.Equal(1, await CountAsync());
    }

    [Fact]
    public async Task ChecksRegisteredDocumentsBeyondTheFirstLookupBatch()
    {
        var rows = Enumerable.Range(1, 1500)
            .Select(i => new object[] { $"Supplier {i}", i.ToString("00000000000000") })
            .ToArray();
        await factory.CreateSupplierAsync(1400.ToString("00000000000000"));

        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers, rows));

        var errors = await ErrorsAsync(response);
        Assert.Equal(["Row 1401"], errors.Keys);
    }

    [Fact]
    public async Task ImportsLargeSpreadsheet()
    {
        var rows = Enumerable.Range(1, 1500)
            .Select(i => new object[] { $"Supplier {i}", i.ToString("00000000000000") })
            .ToArray();

        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers, rows));

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.Equal(1500, await CountAsync());
    }

    [Fact]
    public async Task RejectsMissingColumn()
    {
        var response = await client.PostAsync(Route, Spreadsheet.Upload(["Name"], ["Acme Ltda"]));

        var errors = await ErrorsAsync(response);
        Assert.Contains("Document", Assert.Single(errors["File"]));
    }

    [Fact]
    public async Task RejectsSpreadsheetWithoutRows()
    {
        var response = await client.PostAsync(Route, Spreadsheet.Upload(Headers));

        var errors = await ErrorsAsync(response);
        Assert.Equal(["File"], errors.Keys);
    }

    [Fact]
    public async Task RejectsFileThatIsNotASpreadsheet()
    {
        var response = await client.PostAsync(Route,
            Spreadsheet.Upload(Encoding.UTF8.GetBytes("not a spreadsheet"), "suppliers.xlsx"));

        var errors = await ErrorsAsync(response);
        Assert.Equal(["File"], errors.Keys);
    }

    [Fact]
    public async Task RejectsOtherExtensions()
    {
        var content = Spreadsheet.Build(Headers, ["Acme Ltda", "12345678000195"]);

        var response = await client.PostAsync(Route, Spreadsheet.Upload(content, "suppliers.csv"));

        var errors = await ErrorsAsync(response);
        Assert.Equal(["File"], errors.Keys);
    }

    [Fact]
    public async Task RejectsRequestWithoutFile()
    {
        var form = new MultipartFormDataContent { { new StringContent("value"), "other" } };

        var response = await client.PostAsync(Route, form);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    private static async Task<IDictionary<string, string[]>> ErrorsAsync(HttpResponseMessage response)
    {
        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<HttpValidationProblemDetails>(Json);
        Assert.NotNull(problem);
        return problem.Errors;
    }

    private async Task<int> CountAsync()
    {
        await using var scope = factory.Services.CreateAsyncScope();
        return await scope.ServiceProvider.GetRequiredService<FinanceDbContext>().Suppliers.CountAsync();
    }
}
