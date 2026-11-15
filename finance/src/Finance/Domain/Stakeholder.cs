namespace Finance.Domain;

public class Stakeholder
{
    public int Id { get; private set; }
    public string Name { get; private set; }
    public string Document { get; private set; }
    public StakeholderType Type { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;

    public Stakeholder(string name, string document, StakeholderType type)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);

        if (!TaxDocument.IsValidFormat(document))
        {
            throw new ArgumentException("Document must be a CPF (11 digits) or a CNPJ (14 digits).", nameof(document));
        }

        if (!Enum.IsDefined(type))
        {
            throw new ArgumentOutOfRangeException(nameof(type), type, "Type must be Customer or Supplier.");
        }

        Name = name;
        Document = document;
        Type = type;
    }
}
