-- CreateTable
CREATE TABLE "LedgerEntry" (
    "id" INTEGER NOT NULL,
    "account_id" TEXT NOT NULL,
    "account_type" INTEGER NOT NULL,
    "asset" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL,
    "credit" DECIMAL(18,8) NOT NULL,
    "debit" DECIMAL(18,8) NOT NULL,
    "participant" TEXT NOT NULL,

    CONSTRAINT "LedgerEntry_pkey" PRIMARY KEY ("id")
);
