using Finance.Domain;

namespace Finance.Features.Stakeholders;

public record StakeholderDto(
    int Id,
    string Name,
    string Document,
    StakeholderType Type,
    DateTimeOffset CreatedAt);
