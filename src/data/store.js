const crypto = require('crypto');
const { DEFAULT_COMMISSION_PCT } = require('../config/constants');
const { sha256 } = require('../utils/hash');

// ── In-Memory Datastores ──────────────────────────────────────────────────
const users = {
    'user_customer': {
        id: 'user_customer',
        name: 'Alice Customer',
        role: 'CUSTOMER',
        digitalWallet: {
            fiatBalanceEUR: 50.00,
            loyaltyPoints: 120,
            cashbackRatePct: 5
        },
        cryptoWalletAddress: '0xAlice_Customer_MLDSA65'
    },
    'user_restaurant': {
        id: 'user_restaurant',
        name: 'Napoli Woodfire Pizza',
        role: 'RESTAURANT',
        digitalWallet: {
            fiatBalanceEUR: 1250.00,
            loyaltyPoints: 0
        },
        cryptoWalletAddress: '0xNapoli_Restaurant_MLDSA65'
    },
    'user_courier': {
        id: 'user_courier',
        name: 'Bob Delivery Rider',
        role: 'COURIER',
        digitalWallet: {
            fiatBalanceEUR: 85.50,
            loyaltyPoints: 40
        },
        cryptoWalletAddress: '0xBob_Rider_MLDSA65'
    }
};

const restaurants = {
    'user_restaurant': {
        id: 'user_restaurant',
        name: 'Napoli Woodfire Pizza',
        tagline: 'Authentic Neapolitan Pizza & Artisan Italian Delicacies',
        cuisine: 'Italian, Pizza, Artisan',
        rating: 4.9,
        reviewCount: 342,
        address: 'Via Toledo 42, Napoli / Historic District',
        icon: '🍕',
        isOpen: true,
        prepEtaMinutes: 15,
        deliveryFeeEUR: 2.50,
        minOrderEUR: 15.00,
        deliveryRadiusKm: 5.5,
        cryptoWalletAddress: '0xNapoli_Restaurant_MLDSA65',
        fiatBalanceEUR: 1250.00
    },
    'rest_burger': {
        id: 'rest_burger',
        name: 'Smash & Co. Gourmet Burgers',
        tagline: 'Double Smashed Dry-Aged Angus & Loaded Brioche',
        cuisine: 'American, Burgers, Craft',
        rating: 4.8,
        reviewCount: 218,
        address: 'Corso Umberto I 118, Napoli',
        icon: '🍔',
        isOpen: true,
        prepEtaMinutes: 20,
        deliveryFeeEUR: 2.99,
        minOrderEUR: 12.00,
        deliveryRadiusKm: 6.0,
        cryptoWalletAddress: '0xSmashBurger_MLDSA65',
        fiatBalanceEUR: 890.00
    },
    'rest_sushi': {
        id: 'rest_sushi',
        name: 'Tokyo Bloom Omakase & Bento',
        tagline: 'Sustainably Sourced Sashimi, Nigiri & Crispy Gyoza',
        cuisine: 'Japanese, Sushi, Asian',
        rating: 4.9,
        reviewCount: 451,
        address: 'Via Chiaia 84, Napoli',
        icon: '🍣',
        isOpen: true,
        prepEtaMinutes: 25,
        deliveryFeeEUR: 3.50,
        minOrderEUR: 20.00,
        deliveryRadiusKm: 7.0,
        cryptoWalletAddress: '0xTokyoBloom_MLDSA65',
        fiatBalanceEUR: 2140.00
    },
    'rest_green': {
        id: 'rest_green',
        name: 'Verde Organic Bowls & Cold Press',
        tagline: 'Superfood Macro Bowls, Wild Greens & Fresh Smoothies',
        cuisine: 'Healthy, Vegan, Organic',
        rating: 4.7,
        reviewCount: 129,
        address: 'Piazza dei Martiri 16, Napoli',
        icon: '🥗',
        isOpen: true,
        prepEtaMinutes: 12,
        deliveryFeeEUR: 1.99,
        minOrderEUR: 10.00,
        deliveryRadiusKm: 4.5,
        cryptoWalletAddress: '0xVerdeOrganic_MLDSA65',
        fiatBalanceEUR: 620.00
    }
};

const menus = {
    'user_restaurant': [
        {
            id: 'menu_margherita',
            category: 'Pizza & Mains',
            name: 'Artisanal Margherita Pizza',
            description: 'San Marzano D.O.P. tomatoes, fresh buffalo mozzarella, fragrant basil, extra virgin olive oil.',
            priceEUR: 16.50,
            priceOEN: '1.21',
            inStock: true,
            prepMinutes: 12,
            badge: 'Bestseller'
        },
        {
            id: 'menu_diavola',
            category: 'Pizza & Mains',
            name: 'Spicy Diavola Pizza',
            description: 'Spianata Calabrese spicy salami, smoked provolone, chili flakes, organic tomato reduction.',
            priceEUR: 18.00,
            priceOEN: '1.32',
            inStock: true,
            prepMinutes: 14,
            badge: 'Spicy'
        },
        {
            id: 'menu_arancini',
            category: 'Starters',
            name: 'Truffle & Porcini Arancini',
            description: 'Crispy golden saffron risotto balls stuffed with black truffle cream and melted fontina cheese.',
            priceEUR: 12.00,
            priceOEN: '0.88',
            inStock: true,
            prepMinutes: 8,
            badge: 'Vegetarian'
        },
        {
            id: 'menu_burrata',
            category: 'Starters',
            name: 'Pugliese Burrata & Heirloom Salad',
            description: 'Creamy artisanal burrata, heirloom cherry tomatoes, aged balsamic glaze, toasted pine nuts.',
            priceEUR: 14.00,
            priceOEN: '1.03',
            inStock: true,
            prepMinutes: 6,
            badge: 'Chef Special'
        },
        {
            id: 'menu_tiramisu',
            category: 'Desserts',
            name: 'Classic Espresso Tiramisù',
            description: 'Layered savoiardi soaked in single-origin espresso and Marsala, mascarpone cream, dark cocoa.',
            priceEUR: 8.50,
            priceOEN: '0.62',
            inStock: true,
            prepMinutes: 5,
            badge: 'Homemade'
        },
        {
            id: 'menu_sanpellegrino',
            category: 'Beverages',
            name: 'San Pellegrino Blood Orange',
            description: 'Sparkling Italian citrus beverage crafted with sun-ripened Sicilian blood oranges.',
            priceEUR: 4.50,
            priceOEN: '0.33',
            inStock: true,
            prepMinutes: 2,
            badge: 'Cold'
        }
    ],
    'rest_burger': [
        {
            id: 'menu_smash_truffle',
            category: 'Burgers',
            name: 'Double Truffle Smash Burger',
            description: 'Two 100g dry-aged Angus patties, double American cheese, black truffle aioli, grilled onions, brioche.',
            priceEUR: 15.50,
            priceOEN: '1.14',
            inStock: true,
            prepMinutes: 12,
            badge: 'Bestseller'
        },
        {
            id: 'menu_bacon_bbq',
            category: 'Burgers',
            name: 'Smoked Bacon BBQ Burger',
            description: 'Crispy applewood smoked bacon, aged cheddar, bourbon BBQ glaze, house pickles.',
            priceEUR: 14.50,
            priceOEN: '1.07',
            inStock: true,
            prepMinutes: 10,
            badge: 'Popular'
        },
        {
            id: 'menu_parm_fries',
            category: 'Sides',
            name: 'Parmesan & Rosemary Fries',
            description: 'Triple-cooked rustic skin-on potatoes dusted with 24-month Parmigiano Reggiano and fresh rosemary.',
            priceEUR: 6.00,
            priceOEN: '0.44',
            inStock: true,
            prepMinutes: 6,
            badge: 'Crispy'
        }
    ],
    'rest_sushi': [
        {
            id: 'menu_omakase_set',
            category: 'Nigiri & Sets',
            name: 'Chef Omakase 12-Piece Set',
            description: 'Otoro fatty tuna, Scottish salmon, yellowtail hamachi, Hokkaido scallop, sweet unagi eel.',
            priceEUR: 28.00,
            priceOEN: '2.06',
            inStock: true,
            prepMinutes: 18,
            badge: 'Chef Special'
        },
        {
            id: 'menu_dragon_roll',
            category: 'Special Rolls',
            name: 'Crunchy Tiger Dragon Roll',
            description: 'Tempura black tiger prawn, Hass avocado, cucumber, unagi reduction, flying fish tobiko.',
            priceEUR: 16.00,
            priceOEN: '1.18',
            inStock: true,
            prepMinutes: 14,
            badge: 'Bestseller'
        },
        {
            id: 'menu_gyoza',
            category: 'Starters',
            name: 'Crispy Duck & Ponzu Gyoza',
            description: 'Pan-seared handmade dumplings filled with confit duck and ginger scallion, citrus ponzu dip.',
            priceEUR: 9.50,
            priceOEN: '0.70',
            inStock: true,
            prepMinutes: 8,
            badge: 'Hot'
        }
    ],
    'rest_green': [
        {
            id: 'menu_bali_bowl',
            category: 'Superfood Bowls',
            name: 'Bali Sunset Macro Bowl',
            description: 'Tri-color organic quinoa, Hass avocado, edamame, roasted sweet potato, toasted nori, tahini miso.',
            priceEUR: 13.50,
            priceOEN: '0.99',
            inStock: true,
            prepMinutes: 8,
            badge: 'Organic'
        },
        {
            id: 'menu_green_cleanse',
            category: 'Cold Press',
            name: 'Immunity Cold Press Elixir',
            description: 'Cold-pressed Tuscan kale, Granny Smith apple, fresh ginger root, Meyer lemon, organic celery.',
            priceEUR: 6.50,
            priceOEN: '0.48',
            inStock: true,
            prepMinutes: 3,
            badge: 'Detox'
        }
    ]
};

const orders = {};

// Seed demo order
const seedOrder = {
    id: 'ord_seed_101',
    buyerId: 'user_customer',
    buyerName: 'Alice Customer',
    buyerAddress: '0xAlice_Customer_MLDSA65',
    restaurantId: 'user_restaurant',
    restaurantAddress: '0xNapoli_Restaurant_MLDSA65',
    courierId: null,
    courierAddress: null,
    items: [
        { name: 'Artisanal Margherita Pizza', qty: 1, price: 16.50 },
        { name: 'Truffle & Porcini Arancini', qty: 1, price: 12.00 }
    ],
    amount: 28.50,
    deliveryFee: 3.50,
    tip: 2.00,
    total: 34.00,
    commissionPct: 5,
    paymentMethod: 'CREDIT_CARD',
    cardPayment: { brand: 'Visa', last4: '4242', transactionId: 'txn_seed_4242' },
    paymentStatus: 'PAID',
    status: 'AWAITING_RESTAURANT',
    pickupBarcode: 'PKG-DEMO01',
    pickupBarcodeHash: sha256('PKG-DEMO01'),
    deliveryPin: '4821',
    deliveryPinHash: sha256('4821'),
    createdAt: new Date(Date.now() - 4 * 60000).toISOString(),
    escrowLocked: true
};
orders[seedOrder.id] = seedOrder;

const savedCards = {
    'user_customer': [
        {
            id: 'card_default_visa',
            brand: 'Visa',
            last4: '4242',
            expMonth: 12,
            expYear: 2028,
            cardHolder: 'Alice Customer',
            isDefault: true
        }
    ]
};

const commissionState = {
    defaultCommissionPct: DEFAULT_COMMISSION_PCT
};

module.exports = {
    users,
    restaurants,
    menus,
    orders,
    savedCards,
    commissionState
};
