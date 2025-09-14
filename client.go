package transip

import (
	"context"
	"time"

	"github.com/libdns/libdns"
	"github.com/transip/gotransip/v6"
	transipdomain "github.com/transip/gotransip/v6/domain"
)

func (p *Provider) setupRepository() error {
	if p.repository == nil {
		client, err := gotransip.NewClient(gotransip.ClientConfiguration{
			AccountName:	p.AccountName,
			PrivateKeyPath:	p.PrivateKeyPath,
		})
		if err != nil {
			return err
		}
		p.repository = &transipdomain.Repository{Client: client}
	}

	return nil
}

func (p *Provider) RRToRecord(r libdns.RR) libdns.Record {
	record, err := libdns.RR.Parse(r)
	if err != nil {
		return nil
	}
	return record
}

func (p *Provider) getDNSEntries(ctx context.Context, domain string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return nil, err
	}
	
	var records []libdns.Record
	var dnsEntries []transipdomain.DNSEntry

	dnsEntries, err = p.repository.GetDNSEntries(domain)
	if err != nil {
		return nil, err
	}

	for _, entry := range dnsEntries {
		record := libdns.RR{
			Name:  entry.Name,
			Data: entry.Content,
			Type:  entry.Type,
			TTL:   time.Duration(entry.Expire) * time.Second,
		}
		records = append(records, p.RRToRecord(record))
	}

	return records, nil
}

func (p *Provider) addDNSEntry(ctx context.Context, domain string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return libdns.Record{}, err
	}

	rr := record.RR()
	
	entry := transipdomain.DNSEntry{
		Name:    rr.Name,
		Content: rr.Data,
		Type:    rr.Type,
		Expire:  int(rr.TTL.Seconds()),
	}

	err = p.repository.AddDNSEntry(domain, entry)
	if err != nil {
		return libdns.Record{}, err
	}

	return record, nil
}

func (p *Provider) removeDNSEntry(ctx context.Context, domain string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return libdns.Record{}, err
	}

	entry := transipdomain.DNSEntry{
		Name:    record.Name,
		Content: record.Data,
		Type:    record.Type,
		Expire:  int(record.TTL.Seconds()),
	}

	err = p.repository.RemoveDNSEntry(domain, entry)
	if err != nil {
		return libdns.Record{}, err
	}

	return record, nil
}

func (p *Provider) updateDNSEntry(ctx context.Context, domain string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return libdns.Record{}, err
	}

	entry := transipdomain.DNSEntry{
		Name:    record.Name,
		Content: record.Data,
		Type:    record.Type,
		Expire:  int(record.TTL.Seconds()),
	}

	err = p.repository.UpdateDNSEntry(domain, entry)
	if err != nil {
		return libdns.Record{}, err
	}

	return record, nil
}
