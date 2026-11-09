namespace Finance.Domain;

public class Customer
{
    public int Id { get; private set; }
    public string Name { get; private set; }
    public string Document { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;

    public Customer(string name, string document)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);

        if (!TaxDocument.IsValidFormat(document))
        {
            throw new ArgumentException("Document must be a CPF (11 digits) or a CNPJ (14 digits).", nameof(document));
        }

        Name = name;
        Document = document;
    }
}
