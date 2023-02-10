-- +migrate Down
DROP TABLE IF EXISTS state.monitored_txs;

ALTER TABLE state.verified_batch
DROP COLUMN IF EXISTS is_trusted;

-- +migrate Up
CREATE TABLE state.monitored_txs
(
    owner      VARCHAR NOT NULL,
    id         VARCHAR NOT NULL,
    from_addr  VARCHAR NOT NULL,
    to_addr    VARCHAR,
    nonce      DECIMAL(78, 0) NOT NULL,
    value      DECIMAL(78, 0),
    data       VARCHAR,
    gas        DECIMAL(78, 0) NOT NULL,
    gas_price  DECIMAL(78, 0) NOT NULL,
    status     VARCHAR NOT NULL,
    history    VARCHAR[],
    block_num  BIGINT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (owner, id)
);

ALTER TABLE state.verified_batch
ADD COLUMN is_trusted BOOLEAN DEFAULT true;
