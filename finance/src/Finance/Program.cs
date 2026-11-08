using System.Text.Json.Serialization;
using Finance.Data;
using Finance.Features.BankAccounts;
using Finance.Features.Payables;
using Finance.Features.Receivables;
using Microsoft.EntityFrameworkCore;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddOpenApi();
builder.Services.AddProblemDetails();
builder.Services.AddValidation();
builder.Services.ConfigureHttpJsonOptions(options =>
    options.SerializerOptions.Converters.Add(new JsonStringEnumConverter()));
builder.Services.AddDbContext<FinanceDbContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("Finance")));

var app = builder.Build();

if (app.Environment.IsDevelopment())
{
    app.MapOpenApi();
}

app.UseExceptionHandler(new ExceptionHandlerOptions
{
    StatusCodeSelector = exception => exception is BadHttpRequestException badRequest
        ? badRequest.StatusCode
        : StatusCodes.Status500InternalServerError
});
app.UseStatusCodePages();

app.MapReceivables();
app.MapPayables();
app.MapBankAccounts();

app.Run();
