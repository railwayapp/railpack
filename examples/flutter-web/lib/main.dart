import 'package:flutter/material.dart';

void main() {
  runApp(const RailpackApp());
}

class RailpackApp extends StatelessWidget {
  const RailpackApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Railpack Flutter Web',
      home: Scaffold(
        body: Center(
          child: Text(
            'Hello from Flutter Web!',
            style: Theme.of(context).textTheme.headlineMedium,
          ),
        ),
      ),
    );
  }
}
