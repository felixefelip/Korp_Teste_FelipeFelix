namespace Finance.Domain;

public static class TaxDocument
{
    public const string Pattern = @"^(\d{11}|\d{14})$";

    public static bool IsValidFormat(string? document) =>
        document is { Length: 11 or 14 } && document.All(char.IsAsciiDigit);
}
