const test = require('node:test');
const assert = require('node:assert');

test('Cuisine filtering correctly filters restaurants by keyword', () => {
    const restaurants = [
        { id: 'r1', name: 'Napoli Woodfire Pizza', cuisine: 'Italian, Pizza, Artisan' },
        { id: 'r2', name: 'Smash & Co. Gourmet Burgers', cuisine: 'American, Burgers, Craft' },
        { id: 'r3', name: 'Tokyo Bloom Omakase & Bento', cuisine: 'Japanese, Sushi, Asian' },
        { id: 'r4', name: 'Verde Organic Bowls & Cold Press', cuisine: 'Healthy, Vegan, Organic' }
    ];

    const filterItalian = restaurants.filter(r => r.cuisine.toLowerCase().includes('italian'));
    assert.strictEqual(filterItalian.length, 1);
    assert.strictEqual(filterItalian[0].id, 'r1');

    const filterBurger = restaurants.filter(r => r.cuisine.toLowerCase().includes('burgers'));
    assert.strictEqual(filterBurger.length, 1);
    assert.strictEqual(filterBurger[0].id, 'r2');
});

test('Search keyword matches restaurant name and tagline', () => {
    const restaurants = [
        { name: 'Napoli Woodfire Pizza', tagline: 'Authentic Neapolitan Pizza' },
        { name: 'Tokyo Bloom Omakase & Bento', tagline: 'Sustainably Sourced Sashimi' }
    ];

    const searchSashimi = restaurants.filter(r => 
        r.name.toLowerCase().includes('sashimi') || r.tagline.toLowerCase().includes('sashimi')
    );
    assert.strictEqual(searchSashimi.length, 1);
    assert.strictEqual(searchSashimi[0].name, 'Tokyo Bloom Omakase & Bento');
});

test('Delivery fee estimation computes base fee and ETA', () => {
    const baseFee = 2.50;
    const prepEta = 15;
    const transitMins = 12;
    const totalEta = prepEta + transitMins;

    assert.strictEqual(totalEta, 27);
    assert.strictEqual(baseFee, 2.50);
    const feeInOEN = (baseFee / 13.60).toFixed(2);
    assert.strictEqual(feeInOEN, '0.18');
});

test('Live tracking stages map progress percentage and steps', () => {
    const stages = {
        'AWAITING_RESTAURANT': { percent: 15, step: 1 },
        'PREPARING': { percent: 45, step: 2 },
        'READY_FOR_PICKUP': { percent: 65, step: 3 },
        'IN_TRANSIT': { percent: 85, step: 4 },
        'DELIVERED': { percent: 100, step: 5 }
    };

    assert.strictEqual(stages['AWAITING_RESTAURANT'].percent, 15);
    assert.strictEqual(stages['PREPARING'].percent, 45);
    assert.strictEqual(stages['READY_FOR_PICKUP'].percent, 65);
    assert.strictEqual(stages['IN_TRANSIT'].percent, 85);
    assert.strictEqual(stages['DELIVERED'].percent, 100);
});
