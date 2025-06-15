-- Create "items" table
CREATE TABLE [items] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  CONSTRAINT [PK__items__3213E83F6F35A0A0] PRIMARY KEY CLUSTERED ([id] ASC)
);
-- Create "orders" table
CREATE TABLE [orders] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  CONSTRAINT [PK__orders__3213E83FCD2EB96B] PRIMARY KEY CLUSTERED ([id] ASC)
);
-- Create "order_to_items" table
CREATE TABLE [order_to_items] (
  [order_id] bigint NOT NULL,
  [item_id] bigint NOT NULL,
  CONSTRAINT [PK__order_to__837942D43EDB5071] PRIMARY KEY CLUSTERED ([order_id] ASC, [item_id] ASC),
  CONSTRAINT [FK__order_to___item___24B26D99] FOREIGN KEY ([item_id]) REFERENCES [items] ([id]) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT [FK__order_to___order__25A691D2] FOREIGN KEY ([order_id]) REFERENCES [orders] ([id]) ON UPDATE NO ACTION ON DELETE NO ACTION
);
