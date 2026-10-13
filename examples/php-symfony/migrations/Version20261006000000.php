<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20261006000000 extends AbstractMigration
{
    public function up(Schema $schema): void
    {
        $this->addSql('CREATE TABLE welcome (message VARCHAR(255) NOT NULL)');
        $this->addSql("INSERT INTO welcome (message) VALUES ('Welcome to Symfony')");
    }

    public function down(Schema $schema): void
    {
        $this->addSql('DROP TABLE welcome');
    }
}
