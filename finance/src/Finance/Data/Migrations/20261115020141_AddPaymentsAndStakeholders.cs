using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Finance.Data.Migrations
{
    /// <inheritdoc />
    public partial class AddPaymentsAndStakeholders : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "CustomerName",
                table: "Receivables");

            migrationBuilder.DropColumn(
                name: "SupplierName",
                table: "Payables");

            migrationBuilder.AddColumn<int>(
                name: "CustomerId",
                table: "Receivables",
                type: "int",
                nullable: false,
                defaultValue: 0);

            migrationBuilder.AddColumn<byte[]>(
                name: "RowVersion",
                table: "Receivables",
                type: "rowversion",
                rowVersion: true,
                nullable: true);

            migrationBuilder.AddColumn<byte[]>(
                name: "RowVersion",
                table: "Payables",
                type: "rowversion",
                rowVersion: true,
                nullable: true);

            migrationBuilder.AddColumn<int>(
                name: "SupplierId",
                table: "Payables",
                type: "int",
                nullable: false,
                defaultValue: 0);

            migrationBuilder.CreateTable(
                name: "PayablePayments",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    PayableId = table.Column<int>(type: "int", nullable: false),
                    Amount = table.Column<decimal>(type: "decimal(18,2)", precision: 18, scale: 2, nullable: false),
                    PaidOn = table.Column<DateOnly>(type: "date", nullable: false),
                    CreatedAt = table.Column<DateTimeOffset>(type: "datetimeoffset", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PayablePayments", x => x.Id);
                    table.CheckConstraint("CK_PayablePayments_Amount", "[Amount] > 0");
                    table.ForeignKey(
                        name: "FK_PayablePayments_Payables_PayableId",
                        column: x => x.PayableId,
                        principalTable: "Payables",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "ReceivablePayments",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    ReceivableId = table.Column<int>(type: "int", nullable: false),
                    Amount = table.Column<decimal>(type: "decimal(18,2)", precision: 18, scale: 2, nullable: false),
                    PaidOn = table.Column<DateOnly>(type: "date", nullable: false),
                    CreatedAt = table.Column<DateTimeOffset>(type: "datetimeoffset", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ReceivablePayments", x => x.Id);
                    table.CheckConstraint("CK_ReceivablePayments_Amount", "[Amount] > 0");
                    table.ForeignKey(
                        name: "FK_ReceivablePayments_Receivables_ReceivableId",
                        column: x => x.ReceivableId,
                        principalTable: "Receivables",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "Stakeholders",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Name = table.Column<string>(type: "nvarchar(150)", maxLength: 150, nullable: false),
                    Document = table.Column<string>(type: "varchar(14)", unicode: false, maxLength: 14, nullable: false),
                    Type = table.Column<string>(type: "nvarchar(20)", maxLength: 20, nullable: false),
                    CreatedAt = table.Column<DateTimeOffset>(type: "datetimeoffset", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Stakeholders", x => x.Id);
                });

            migrationBuilder.CreateIndex(
                name: "IX_Receivables_CustomerId",
                table: "Receivables",
                column: "CustomerId");

            migrationBuilder.CreateIndex(
                name: "IX_Payables_SupplierId",
                table: "Payables",
                column: "SupplierId");

            migrationBuilder.CreateIndex(
                name: "IX_PayablePayments_PayableId",
                table: "PayablePayments",
                column: "PayableId");

            migrationBuilder.CreateIndex(
                name: "IX_ReceivablePayments_ReceivableId",
                table: "ReceivablePayments",
                column: "ReceivableId");

            migrationBuilder.CreateIndex(
                name: "IX_Stakeholders_Type_Document",
                table: "Stakeholders",
                columns: new[] { "Type", "Document" },
                unique: true);

            migrationBuilder.AddForeignKey(
                name: "FK_Payables_Stakeholders_SupplierId",
                table: "Payables",
                column: "SupplierId",
                principalTable: "Stakeholders",
                principalColumn: "Id",
                onDelete: ReferentialAction.Restrict);

            migrationBuilder.AddForeignKey(
                name: "FK_Receivables_Stakeholders_CustomerId",
                table: "Receivables",
                column: "CustomerId",
                principalTable: "Stakeholders",
                principalColumn: "Id",
                onDelete: ReferentialAction.Restrict);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_Payables_Stakeholders_SupplierId",
                table: "Payables");

            migrationBuilder.DropForeignKey(
                name: "FK_Receivables_Stakeholders_CustomerId",
                table: "Receivables");

            migrationBuilder.DropTable(
                name: "PayablePayments");

            migrationBuilder.DropTable(
                name: "ReceivablePayments");

            migrationBuilder.DropTable(
                name: "Stakeholders");

            migrationBuilder.DropIndex(
                name: "IX_Receivables_CustomerId",
                table: "Receivables");

            migrationBuilder.DropIndex(
                name: "IX_Payables_SupplierId",
                table: "Payables");

            migrationBuilder.DropColumn(
                name: "CustomerId",
                table: "Receivables");

            migrationBuilder.DropColumn(
                name: "RowVersion",
                table: "Receivables");

            migrationBuilder.DropColumn(
                name: "RowVersion",
                table: "Payables");

            migrationBuilder.DropColumn(
                name: "SupplierId",
                table: "Payables");

            migrationBuilder.AddColumn<string>(
                name: "CustomerName",
                table: "Receivables",
                type: "nvarchar(150)",
                maxLength: 150,
                nullable: false,
                defaultValue: "");

            migrationBuilder.AddColumn<string>(
                name: "SupplierName",
                table: "Payables",
                type: "nvarchar(150)",
                maxLength: 150,
                nullable: false,
                defaultValue: "");
        }
    }
}
