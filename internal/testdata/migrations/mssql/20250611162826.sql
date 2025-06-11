-- Create "users" table
CREATE TABLE [users] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  [name] varchar(255) NULL,
  [emails] varchar(255) NULL,
  CONSTRAINT [PK__users__3213E83F244477B4] PRIMARY KEY CLUSTERED ([id] ASC)
);
-- Create "stories" table
CREATE TABLE [stories] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  [title] varchar(255) NULL,
  [author_id] bigint NULL,
  CONSTRAINT [PK__stories__3213E83F8ECA653E] PRIMARY KEY CLUSTERED ([id] ASC),
  CONSTRAINT [FK__stories__author___22CA2527] FOREIGN KEY ([author_id]) REFERENCES [users] ([id]) ON UPDATE NO ACTION ON DELETE NO ACTION
);
