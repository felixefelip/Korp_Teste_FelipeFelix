using System.ComponentModel.DataAnnotations;
using Finance.Domain;

namespace Finance.Features.Stakeholders;

public record CreateStakeholderDto(
    [property: Required, MaxLength(150)] string Name,
    [property: Required, RegularExpression(TaxDocument.Pattern)] string Document,
    [property: Required, EnumDataType(typeof(StakeholderType))] StakeholderType? Type);
