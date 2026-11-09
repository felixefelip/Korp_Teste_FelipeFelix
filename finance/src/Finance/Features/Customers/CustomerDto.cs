namespace Finance.Features.Customers;

public record CustomerDto(
    int Id,
    string Name,
    string Document,
    DateTimeOffset CreatedAt);
