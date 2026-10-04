const { restaurants, menus, orders, commissionState } = require('../data/store');
const { OEN_EUR_EXCHANGE_RATE, LEGACY_APP_FEE_PCT } = require('../config/constants');

function listRestaurants({ cuisine, search }) {
    let list = Object.values(restaurants);

    if (cuisine && cuisine !== 'ALL') {
        const cLower = cuisine.toLowerCase();
        list = list.filter(r => r.cuisine.toLowerCase().includes(cLower));
    }

    if (search) {
        const sLower = search.toLowerCase();
        list = list.filter(r => 
            r.name.toLowerCase().includes(sLower) || 
            r.cuisine.toLowerCase().includes(sLower) ||
            r.tagline.toLowerCase().includes(sLower)
        );
    }

    return {
        success: true,
        count: list.length,
        restaurants: list
    };
}

function getRestaurantProfile(id) {
    const restaurant = restaurants[id];
    if (!restaurant) {
        const error = new Error('Restaurant not found');
        error.statusCode = 404;
        throw error;
    }

    const defaultCommissionPct = commissionState.defaultCommissionPct;
    const restaurantOrders = Object.values(orders).filter(o => o.restaurantId === restaurant.id);
    const activeOrders = restaurantOrders.filter(o => 
        ['AWAITING_RESTAURANT', 'PREPARING', 'READY_FOR_PICKUP', 'IN_TRANSIT'].includes(o.status)
    );
    const deliveredOrders = restaurantOrders.filter(o => o.status === 'DELIVERED');
    
    const grossDeliveredRevenueEUR = deliveredOrders.reduce((sum, o) => sum + (o.restaurantPayout || (o.amount * 0.95)), 0);
    const activePipelineRevenueEUR = activeOrders.reduce((sum, o) => sum + (o.amount * ((100 - (o.commissionPct || defaultCommissionPct)) / 100)), 0);
    const totalRevenueTodayEUR = parseFloat((grossDeliveredRevenueEUR + activePipelineRevenueEUR).toFixed(2));
    
    const commissionRetainedPct = 100 - defaultCommissionPct;
    const legacyLostRevenueEUR = parseFloat((totalRevenueTodayEUR * (LEGACY_APP_FEE_PCT - defaultCommissionPct) / 100).toFixed(2));

    return {
        success: true,
        restaurant: {
            ...restaurant,
            stats: {
                totalOrdersCount: restaurantOrders.length,
                activeOrdersCount: activeOrders.length,
                deliveredOrdersCount: deliveredOrders.length,
                grossRevenueTodayEUR: totalRevenueTodayEUR,
                fiatBalanceEUR: restaurant.fiatBalanceEUR,
                commissionRetainedPct,
                platformCommissionPct: defaultCommissionPct,
                legacyLostRevenueEUR,
                avgPrepTimeMinutes: restaurant.prepEtaMinutes || 15
            }
        }
    };
}

function updateRestaurantProfile(id, { isOpen, prepEtaMinutes, name, tagline, deliveryRadiusKm }) {
    const restaurant = restaurants[id];
    if (!restaurant) {
        const error = new Error('Restaurant not found');
        error.statusCode = 404;
        throw error;
    }

    if (typeof isOpen === 'boolean') restaurant.isOpen = isOpen;
    if (typeof prepEtaMinutes === 'number') restaurant.prepEtaMinutes = prepEtaMinutes;
    if (name) restaurant.name = name;
    if (tagline) restaurant.tagline = tagline;
    if (typeof deliveryRadiusKm === 'number') restaurant.deliveryRadiusKm = deliveryRadiusKm;

    return {
        success: true,
        message: 'Restaurant profile updated successfully',
        restaurant
    };
}

function getMenu(restaurantId) {
    const menuList = menus[restaurantId] || [];
    return {
        success: true,
        restaurantId,
        count: menuList.length,
        menu: menuList
    };
}

function addDish(restaurantId, { name, category = 'Mains', description = '', priceEUR, prepMinutes = 15, badge = 'New' }) {
    if (!name || !priceEUR) {
        const error = new Error('Dish name and price (EUR) are required');
        error.statusCode = 400;
        throw error;
    }

    if (!menus[restaurantId]) menus[restaurantId] = [];
    const priceNum = parseFloat(priceEUR);
    const newItem = {
        id: `menu_${Date.now()}`,
        category,
        name: name.trim(),
        description: description.trim(),
        priceEUR: priceNum,
        priceOEN: (priceNum / OEN_EUR_EXCHANGE_RATE).toFixed(2),
        inStock: true,
        prepMinutes: parseInt(prepMinutes, 10) || 15,
        badge
    };

    menus[restaurantId].push(newItem);
    return {
        success: true,
        message: `Dish "${newItem.name}" added to catalog`,
        item: newItem
    };
}

function updateDish(restaurantId, itemId, { inStock, priceEUR, name, description, category, prepMinutes, badge }) {
    const menuList = menus[restaurantId];
    if (!menuList) {
        const error = new Error('Restaurant menu not found');
        error.statusCode = 404;
        throw error;
    }

    const item = menuList.find(i => i.id === itemId);
    if (!item) {
        const error = new Error('Menu item not found');
        error.statusCode = 404;
        throw error;
    }

    if (typeof inStock === 'boolean') item.inStock = inStock;
    if (priceEUR !== undefined) {
        const p = parseFloat(priceEUR);
        item.priceEUR = p;
        item.priceOEN = (p / OEN_EUR_EXCHANGE_RATE).toFixed(2);
    }
    if (name) item.name = name.trim();
    if (description !== undefined) item.description = description.trim();
    if (category) item.category = category;
    if (prepMinutes) item.prepMinutes = parseInt(prepMinutes, 10);
    if (badge) item.badge = badge;

    return {
        success: true,
        message: `Dish "${item.name}" updated`,
        item
    };
}

function deleteDish(restaurantId, itemId) {
    const menuList = menus[restaurantId];
    if (!menuList) {
        const error = new Error('Restaurant menu not found');
        error.statusCode = 404;
        throw error;
    }

    const index = menuList.findIndex(i => i.id === itemId);
    if (index === -1) {
        const error = new Error('Menu item not found');
        error.statusCode = 404;
        throw error;
    }

    const removed = menuList.splice(index, 1)[0];
    return {
        success: true,
        message: `Dish "${removed.name}" deleted from menu`,
        deletedItemId: removed.id
    };
}

function getRestaurantOrders(restaurantId, status) {
    let list = Object.values(orders).filter(o => o.restaurantId === restaurantId);

    if (status) {
        const statuses = status.split(',').map(s => s.trim());
        list = list.filter(o => statuses.includes(o.status));
    }

    list.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
    return {
        success: true,
        count: list.length,
        orders: list
    };
}

module.exports = {
    listRestaurants,
    getRestaurantProfile,
    updateRestaurantProfile,
    getMenu,
    addDish,
    updateDish,
    deleteDish,
    getRestaurantOrders
};
