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

-- CreateTable
CREATE TABLE "Channel" (
    "channel_id" TEXT NOT NULL,
    "adjudicator" TEXT NOT NULL,
    "amount" BIGINT NOT NULL,
    "chain_id" INTEGER NOT NULL,
    "challenge" INTEGER NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL,
    "nonce" BIGINT NOT NULL,
    "participant" TEXT NOT NULL,
    "status" TEXT NOT NULL,
    "token" TEXT NOT NULL,
    "updated_at" TIMESTAMP(3) NOT NULL,
    "version" INTEGER NOT NULL,
    "wallet" TEXT NOT NULL,

    CONSTRAINT "Channel_pkey" PRIMARY KEY ("channel_id")
);

-- CreateIndex
CREATE INDEX "Channel_chain_id_idx" ON "Channel"("chain_id");

-- CreateIndex
CREATE INDEX "Channel_token_idx" ON "Channel"("token");
