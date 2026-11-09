using System.ComponentModel.DataAnnotations;
using Finance.Domain;

namespace Finance.Features.Customers;

public record CreateCustomerDto(
    [property: Required, MaxLength(150)] string Name,
    [property: Required, RegularExpression(TaxDocument.Pattern)] string Document);
