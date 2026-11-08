using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Finance.Data.Migrations
{
    /// <inheritdoc />
    public partial class AddPaidAmountAndBankAccounts : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<int>(
                name: "BankAccountId",
                table: "Receivables",
                type: "int",
                nullable: false,
                defaultValue: 0);

            migrationBuilder.AddColumn<decimal>(
                name: "PaidAmount",
                table: "Receivables",
                type: "decimal(18,2)",
                precision: 18,
                scale: 2,
                nullable: false,
                defaultValue: 0m);

            migrationBuilder.AddColumn<int>(
                name: "BankAccountId",
                table: "Payables",
                type: "int",
                nullable: false,
                defaultValue: 0);

            migrationBuilder.AddColumn<decimal>(
                name: "PaidAmount",
                table: "Payables",
                type: "decimal(18,2)",
                precision: 18,
                scale: 2,
                nullable: false,
                defaultValue: 0m);

            migrationBuilder.CreateTable(
                name: "BankAccounts",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Name = table.Column<string>(type: "nvarchar(100)", maxLength: 100, nullable: false),
                    BankCode = table.Column<string>(type: "nchar(3)", fixedLength: true, maxLength: 3, nullable: false),
                    Branch = table.Column<string>(type: "nvarchar(10)", maxLength: 10, nullable: false),
                    AccountNumber = table.Column<string>(type: "nvarchar(20)", maxLength: 20, nullable: false),
                    OpeningBalance = table.Column<decimal>(type: "decimal(18,2)", precision: 18, scale: 2, nullable: false),
                    CreatedAt = table.Column<DateTimeOffset>(type: "datetimeoffset", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_BankAccounts", x => x.Id);
                });

            migrationBuilder.CreateIndex(
                name: "IX_Receivables_BankAccountId",
                table: "Receivables",
                column: "BankAccountId");

            migrationBuilder.AddCheckConstraint(
                name: "CK_Receivables_PaidAmount",
                table: "Receivables",
                sql: "[PaidAmount] >= 0 AND [PaidAmount] <= [Amount]");

            migrationBuilder.CreateIndex(
                name: "IX_Payables_BankAccountId",
                table: "Payables",
                column: "BankAccountId");

            migrationBuilder.AddCheckConstraint(
                name: "CK_Payables_PaidAmount",
                table: "Payables",
                sql: "[PaidAmount] >= 0 AND [PaidAmount] <= [Amount]");

            migrationBuilder.CreateIndex(
                name: "IX_BankAccounts_BankCode_Branch_AccountNumber",
                table: "BankAccounts",
                columns: new[] { "BankCode", "Branch", "AccountNumber" },
                unique: true);

            migrationBuilder.AddForeignKey(
                name: "FK_Payables_BankAccounts_BankAccountId",
                table: "Payables",
                column: "BankAccountId",
                principalTable: "BankAccounts",
                principalColumn: "Id",
                onDelete: ReferentialAction.Restrict);

            migrationBuilder.AddForeignKey(
                name: "FK_Receivables_BankAccounts_BankAccountId",
                table: "Receivables",
                column: "BankAccountId",
                principalTable: "BankAccounts",
                principalColumn: "Id",
                onDelete: ReferentialAction.Restrict);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_Payables_BankAccounts_BankAccountId",
                table: "Payables");

            migrationBuilder.DropForeignKey(
                name: "FK_Receivables_BankAccounts_BankAccountId",
                table: "Receivables");

            migrationBuilder.DropTable(
                name: "BankAccounts");

            migrationBuilder.DropIndex(
                name: "IX_Receivables_BankAccountId",
                table: "Receivables");

            migrationBuilder.DropCheckConstraint(
                name: "CK_Receivables_PaidAmount",
                table: "Receivables");

            migrationBuilder.DropIndex(
                name: "IX_Payables_BankAccountId",
                table: "Payables");

            migrationBuilder.DropCheckConstraint(
                name: "CK_Payables_PaidAmount",
                table: "Payables");

            migrationBuilder.DropColumn(
                name: "BankAccountId",
                table: "Receivables");

            migrationBuilder.DropColumn(
                name: "PaidAmount",
                table: "Receivables");

            migrationBuilder.DropColumn(
                name: "BankAccountId",
                table: "Payables");

            migrationBuilder.DropColumn(
                name: "PaidAmount",
                table: "Payables");
        }
    }
}
