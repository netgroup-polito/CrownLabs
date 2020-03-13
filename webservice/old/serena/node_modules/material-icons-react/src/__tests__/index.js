import React from 'react';

import { shallow, mount, render } from 'enzyme';
import { expect } from 'chai';
import sinon from 'sinon';

import MaterialIcon from '../index';

const loadIcon = (props) => {
    const wrapper = mount(<MaterialIcon {...props} />);
    
    return wrapper;
}

describe('MaterialIcon renders', () => {
    it('with an `i`', () => {
        const wrapper = loadIcon({icon: 'face'})
    
        expect(wrapper.find('i')).to.be.not.null
    });

    it('with the icon prop', () => {
        const wrapper = loadIcon({icon: 'face'})
    
        expect(wrapper.props().icon).to.equal('face');
    });

    it('with class md-24 by default', async () => {
        const wrapper = loadIcon({ icon: 'face' });
        
        expect(wrapper.find('i').hasClass('md-24')).to.equal(true);
    });

    it('with size prop overridden to medium', async () => {
        const wrapper = loadIcon({ icon: 'face', size: 'medium' });
        
        expect(wrapper.find('i').hasClass('md-36')).to.equal(true);
    });

    it('with dark color by default', () => {
        const wrapper = loadIcon({ icon: 'face' });
        expect(wrapper.find('.material-icons').hasClass('md-dark')).to.equal(true);
    });

    it('with light color when inverted', () => {
        const wrapper = loadIcon({ icon: 'face', invert: true });
        expect(wrapper.find('.material-icons').hasClass('md-light')).to.equal(true);
    });

    it('with active state by default', () => {
        const wrapper = loadIcon({ icon: 'face' });
        expect(wrapper.find('.material-icons').hasClass('md-inactive')).to.equal(false);
    });

    it('with `md-inactive` when inactive', () => {
        const wrapper = loadIcon({ icon: 'face', inactive: true });
        expect(wrapper.find('.material-icons').hasClass('md-inactive')).to.equal(true);
    });

    it('with active state by default when inverted', () => {
        const wrapper = loadIcon({ icon: 'face', invert: true });
        expect(wrapper.find('.material-icons').hasClass('md-light')).to.equal(true);
        expect(wrapper.find('.material-icons').hasClass('md-inactive')).to.equal(false);
    });

    it('with `md-inactive` and `md-light` when inactive and inverted', () => {
        const wrapper = loadIcon({ icon: 'face', invert: true, inactive: true });
        expect(wrapper.find('.material-icons').hasClass('md-light')).to.equal(true);
        expect(wrapper.find('.material-icons').hasClass('md-inactive')).to.equal(true);
    });

    it('with default classes overriden when className prop provided', () => {
        const wrapper = loadIcon({ icon: 'face', className: 'test-class-name' });

        expect(wrapper.find('i').hasClass('test-class-name')).to.equal(true);
    });

    it('with custom props available', () => {
        const wrapper = loadIcon({ icon: 'face', testProp: 'custom-prop'});

        expect(wrapper.props().testProp).to.be.not.null
        expect(wrapper.props().testProp).to.equal('custom-prop')
    });
});