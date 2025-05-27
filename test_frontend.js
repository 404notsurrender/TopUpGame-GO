const puppeteer = require('puppeteer');

(async () => {
    console.log('\x1b[34m%s\x1b[0m', 'Starting Frontend Tests...');
    const browser = await puppeteer.launch({ headless: false });
    const page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 800 });

    try {
        // 1. Test Homepage
        console.log('\n=== Testing Homepage ===');
        await page.goto('http://localhost:8080');
        await page.waitForSelector('.product-list');
        console.log('✓ Homepage loaded successfully');

        // 2. Test Registration
        console.log('\n=== Testing Registration ===');
        await page.goto('http://localhost:8080/register');
        await page.type('#email', 'test@example.com');
        await page.type('#password', 'test123');
        await page.type('#confirmPassword', 'test123');
        await page.select('#role', 'reseller');
        await page.click('button[type="submit"]');
        await page.waitForNavigation();
        console.log('✓ Registration completed');

        // 3. Test Login
        console.log('\n=== Testing Login ===');
        await page.goto('http://localhost:8080/login');
        await page.type('#email', 'test@example.com');
        await page.type('#password', 'test123');
        await page.click('button[type="submit"]');
        await page.waitForNavigation();
        console.log('✓ Login successful');

        // 4. Test Product Checkout
        console.log('\n=== Testing Product Checkout ===');
        await page.goto('http://localhost:8080');
        await page.waitForSelector('.product-list');
        await page.click('button.buy-now');
        await page.waitForSelector('#checkoutModal');
        await page.type('#gameId', '12345');
        await page.type('#gameServer', '1001');
        await page.select('#paymentMethod', 'bank_transfer');
        await page.click('#checkoutForm button[type="submit"]');
        await page.waitForSelector('.success-message');
        console.log('✓ Checkout process completed');

        // 5. Test Transaction Status
        console.log('\n=== Testing Transaction Status ===');
        const invoice = await page.$eval('.invoice-number', el => el.textContent);
        await page.goto(`http://localhost:8080/transaction/${invoice}`);
        await page.waitForSelector('.transaction-details');
        console.log('✓ Transaction status page loaded');

        // 6. Test Admin Dashboard
        console.log('\n=== Testing Admin Dashboard ===');
        await page.goto('http://localhost:8080/login');
        await page.type('#email', 'admin@test.com');
        await page.type('#password', 'admin123');
        await page.click('button[type="submit"]');
        await page.waitForNavigation();
        await page.goto('http://localhost:8080/admin');
        await page.waitForSelector('.admin-dashboard');
        
        // Test Product Management
        await page.click('#products-tab');
        await page.waitForSelector('.product-list');
        await page.click('button.add-product');
        await page.waitForSelector('#addProductModal');
        await page.type('input[name="name"]', 'Test Product');
        await page.type('input[name="category"]', 'Test Category');
        await page.type('input[name="price"]', '10000');
        await page.type('input[name="sku"]', 'TEST-001');
        await page.type('textarea[name="description"]', 'Test Description');
        await page.click('#addProductForm button[type="submit"]');
        await page.waitForSelector('.success-message');
        console.log('✓ Product management tested');

        // Test Transaction Management
        await page.click('#transactions-tab');
        await page.waitForSelector('.transaction-list');
        await page.select('#statusFilter', 'pending');
        await page.waitForSelector('.transaction-item');
        console.log('✓ Transaction management tested');

        // 7. Test Error Handling
        console.log('\n=== Testing Error Handling ===');
        
        // Invalid Login
        await page.goto('http://localhost:8080/login');
        await page.type('#email', 'invalid@test.com');
        await page.type('#password', 'wrong');
        await page.click('button[type="submit"]');
        await page.waitForSelector('.error-message');
        console.log('✓ Invalid login error handled');

        // Invalid Checkout
        await page.goto('http://localhost:8080');
        await page.click('button.buy-now');
        await page.waitForSelector('#checkoutModal');
        await page.click('#checkoutForm button[type="submit"]');
        await page.waitForSelector('.validation-error');
        console.log('✓ Invalid checkout error handled');

        console.log('\n\x1b[32m%s\x1b[0m', 'All frontend tests completed successfully!');
    } catch (error) {
        console.error('\x1b[31m%s\x1b[0m', `Test failed: ${error.message}`);
    } finally {
        await browser.close();
    }
})();
