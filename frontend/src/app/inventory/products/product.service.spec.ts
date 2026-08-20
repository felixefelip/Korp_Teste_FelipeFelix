import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { Product } from './product.model';
import { ProductService } from './product.service';

describe('ProductService', () => {
  let service: ProductService;
  let http: HttpTestingController;

  const newProduct = {
    code: 'PRD-0100',
    name: 'Cadeira de escritório',
    unit: 'UN',
    price: 750.5,
    stock: 8
  };

  const products: Product[] = [
    {
      id: 1,
      code: 'PRD-0001',
      name: 'Notebook Dell Inspiron 15',
      unit: 'UN',
      price: 4299.9,
      stock: 12
    },
    {
      id: 2,
      code: 'PRD-0005',
      name: 'Papel Sulfite A4 75g (resma)',
      unit: 'CX',
      price: 27.4,
      stock: 56
    }
  ];

  const load = (list: Product[] = products) => {
    service.list().subscribe();
    http.expectOne('/api/products').flush(list);
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()]
    });

    service = TestBed.inject(ProductService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  describe('list', () => {
    it('starts empty until the first load', () => {
      expect(service.products()).toEqual([]);
    });

    it('fetches the listing from the API endpoint', () => {
      service.list().subscribe();

      const request = http.expectOne('/api/products');
      expect(request.request.method).toBe('GET');

      request.flush(products);
    });

    it('publishes to the signal whatever the API returned', () => {
      load();
      expect(service.products()).toEqual(products);
    });

    it('replaces the previous list on every load', () => {
      load();
      load([products[0]]);

      expect(service.products()).toEqual([products[0]]);
    });

    it('propagates the failure and keeps the previous list', () => {
      load();

      let failed = false;
      service.list().subscribe({ error: () => (failed = true) });
      http.expectOne('/api/products').flush(null, { status: 500, statusText: 'Error' });

      expect(failed).toBe(true);
      expect(service.products()).toEqual(products);
    });
  });

  describe('nextCode', () => {
    it('suggests the sequence after the highest loaded code', () => {
      load();
      expect(service.nextCode()).toBe('PRD-0006');
    });

    it('suggests the first code when there are no products yet', () => {
      expect(service.nextCode()).toBe('PRD-0001');
    });

    it('moves the suggestion forward after a matching creation', () => {
      load();

      service.create({ ...newProduct, code: 'PRD-0009' }).subscribe();
      http.expectOne('/api/products').flush({ ...newProduct, id: 3, code: 'PRD-0009' });

      expect(service.nextCode()).toBe('PRD-0010');
    });

    it('ignores codes outside the PRD-0000 pattern', () => {
      load([...products, { ...newProduct, id: 3, code: 'ABC-9999' }]);
      expect(service.nextCode()).toBe('PRD-0006');
    });
  });

  describe('create', () => {
    it('posts the product and returns what the API created', () => {
      const created = { ...newProduct, id: 7 };
      let received: Product | undefined;

      service.create(newProduct).subscribe((product) => (received = product));

      const request = http.expectOne('/api/products');
      expect(request.request.method).toBe('POST');
      expect(request.request.body).toEqual(newProduct);

      request.flush(created);

      expect(received).toEqual(created);
    });

    it('appends to the listing the product returned by the API', () => {
      load();

      service.create(newProduct).subscribe();
      http.expectOne('/api/products').flush({ ...newProduct, id: 7, code: 'PRD-0100' });

      expect(service.products()).toHaveLength(3);
      expect(service.products().at(-1)).toMatchObject({ id: 7, code: 'PRD-0100' });
    });

    it('does not mutate the previous list array', () => {
      load();
      const previousList = service.products();

      service.create(newProduct).subscribe();
      http.expectOne('/api/products').flush({ ...newProduct, id: 7 });

      expect(service.products()).not.toBe(previousList);
      expect(previousList).toHaveLength(2);
    });

    it('appends nothing when the API rejects the product', () => {
      load();

      let failed = false;
      service.create(newProduct).subscribe({ error: () => (failed = true) });
      http
        .expectOne('/api/products')
        .flush(
          { errors: { code: 'obrigatorio' } },
          { status: 400, statusText: 'Bad Request' }
        );

      expect(failed).toBe(true);
      expect(service.products()).toEqual(products);
    });
  });
});
