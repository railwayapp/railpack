<?php

namespace App\Controller;

use Doctrine\DBAL\Connection;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

class HomeController
{
    #[Route('/', name: 'home')]
    public function index(Connection $connection): Response
    {
        // The HTTP check only succeeds after the startup migration creates and seeds this table.
        return new Response($connection->fetchOne('SELECT message FROM welcome'));
    }
}
