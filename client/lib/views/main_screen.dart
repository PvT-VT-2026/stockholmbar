import 'package:flutter/material.dart';
import 'map_view.dart';

// "Main" screen som visar sidor utifrån navbar
class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  int _selectedIndex = 0;

  // Sidorna definieras här
  // TODO: Ersätt Bidra och Profil med riktiga sidor när de är klara
  final List<Widget> _pages = [
    const MapView(),
    const Center(child: Text('Fota Meny - WIP')),
    const Center(child: Text('Profil - WIP')),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _pages[_selectedIndex],
      bottomNavigationBar: BottomNavigationBar(
        backgroundColor: Colors.white,
        selectedItemColor: Colors.red,
        unselectedItemColor: Colors.grey[800],
        currentIndex: _selectedIndex,
        onTap: (index) {
          setState(() { _selectedIndex = index; });
        },
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.map), label: 'Karta'),
          BottomNavigationBarItem(icon: Icon(Icons.add_a_photo), label: 'Bidra'),
          BottomNavigationBarItem(icon: Icon(Icons.person), label: 'Profil'),
        ],
      ),
    );
  }
}